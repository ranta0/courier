package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/ranta0/courier"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Vars     interface{}       `yaml:"vars"`
	Requests []courier.UseCase `yaml:"requests"`
}

var (
	version                     = "0.2.1"
	defaultConfigFileName       = "courier"
	defaultConfigFileNameFolder = "config"
	defaultConfigFolderName     = ".courier/"
)

func getConfigFile(configFileName string) (string, error) {
	_, err := os.Stat(configFileName)
	if err != nil {
		configFileName = defaultConfigFileName + ".yaml"
	}

	_, err = os.Stat(configFileName)
	if err != nil {
		configFileName = defaultConfigFolderName + defaultConfigFileNameFolder + ".yaml"
	}

	_, err = os.Stat(configFileName)
	if err != nil {
		return "", fmt.Errorf("configuration file not found")
	}

	return configFileName, nil
}

func newConfig(configFileName string) (*Config, error) {
	yamlFile, err := os.ReadFile(configFileName)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func editConfigFile(editor, configFile string) error {
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}

	if editor == "" {
		return fmt.Errorf("EDITOR enviroment variable is empty or no single editor has been chosen")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", editor+" "+configFile)
	} else {
		cmd = exec.Command("sh", "-c", editor+" "+configFile)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func editorReader(editor, content, extension string) error {
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}

	temp, err := os.CreateTemp("", "courier-resp*"+extension)
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	fmt.Printf("%s", temp.Name())
	temp.Write([]byte(content))

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", editor+" "+temp.Name())
	} else {
		cmd = exec.Command("sh", "-c", editor+" "+temp.Name())
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func prettyJSON(jsonString string) (string, error) {
	var output interface{}

	if err := json.Unmarshal([]byte(jsonString), &output); err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func main() {
	var showVersion bool
	var editor string
	var openResponseEditor bool
	var prettierJSON bool
	var configFile string
	var editFile bool
	var test bool

	flag.StringVar(&editor, "editor", "", "pick the editor, by default it uses the enviroment $EDITOR")
	flag.BoolVar(&openResponseEditor, "open", false, "open each response inside the editor of preference")
	flag.BoolVar(&prettierJSON, "json", false, "prettier json response")
	flag.StringVar(&configFile, "config-file", "", "request file, defaults to courier.yaml or .courier/config.yaml")
	flag.BoolVar(&editFile, "edit", false, "uses editor to edit config file")
	flag.BoolVar(&test, "test", false, "show testing response")
	flag.BoolVar(&showVersion, "v", false, "print version")
	flag.Usage = func() {
		fmt.Printf("%s - simple http client\n\nVersion: %s (%s)\n\nOptions:\n",
			defaultConfigFileName, version, runtime.Version())
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("%s - simple http client\n\nVersion: %s (%s)\n\n",
			defaultConfigFileName, version, runtime.Version())
		os.Exit(0)
	}

	configFileClean, err := getConfigFile(configFile)
	if err != nil {
		fmt.Printf("%s: %s\n", courier.Red("Error"), err)
		os.Exit(1)
	}

	if editFile {
		err = editConfigFile(editor, configFileClean)
		if err != nil {
			fmt.Printf("%s: %s\n", courier.Red("Error"), err)
			os.Exit(1)
		}
	}

	config, err := newConfig(configFileClean)
	if err != nil {
		fmt.Printf("%s: %s\n", courier.Red("Error"), err)
		os.Exit(1)
	}

	for _, value := range config.Requests {
		usecase, err := courier.NewAPIUseCase(config.Vars, &value)
		if err != nil {
			fmt.Printf("%s: %s %s\n", courier.Red("Error"), courier.Blue(usecase.Prefix()), err)
			os.Exit(1)
		}

		var responseOutput string
		if test {
			err = usecase.Test(config.Vars)
			if err != nil {
				fmt.Printf("%s: %s %s\n", courier.Red("Error"), courier.Blue(usecase.Prefix()), err)
			} else {
				fmt.Printf("%s: %s\n", courier.Green("Success"), courier.Blue(usecase.Prefix()))
			}

			continue
		}

		responseOutput, err = usecase.Curl(config.Vars)
		if err != nil {
			fmt.Printf("%s", responseOutput)
			fmt.Printf("%s: %s %s\n", courier.Red("Error"), courier.Blue(usecase.Prefix()), err)
			os.Exit(1)
		}

		extension := ""
		if prettierJSON {
			responseOutput, err = prettyJSON(responseOutput)
			if err != nil {
				fmt.Printf("%s: %s %s\n", courier.Red("Error"), courier.Blue(usecase.Prefix()), err)
				os.Exit(1)
			}

			extension = ".json"
		}

		if openResponseEditor {
			editorReader(editor, responseOutput, extension)
		} else {
			fmt.Printf("%s", responseOutput)
		}
	}
}
