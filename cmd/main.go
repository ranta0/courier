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
	Auth     courier.UseCase   `yaml:"auth"`
	Requests []courier.UseCase `yaml:"requests"`
}

var (
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
	var editor string
	var openResponseEditor bool
	var prettierJson bool
	var configFile string
	var editFile bool
	var test bool

	flag.StringVar(&editor, "e", "", "pick the editor, by default it uses the enviroment $EDITOR")
	flag.BoolVar(&openResponseEditor, "o", false, "open each response inside the editor of preference")
	flag.BoolVar(&prettierJson, "json", false, "prettier json response")
	flag.StringVar(&configFile, "f", "", "request file, defaults to courier.yaml or .courier/config.yaml")
	flag.BoolVar(&editFile, "edit", false, "uses editor to edit config file")
	flag.BoolVar(&test, "t", false, "show testing response")
	flag.Parse()

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
		if prettierJson {
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
