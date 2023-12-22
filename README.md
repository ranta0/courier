# courier
Http client to make testing apis, or just making normal curl requests a bit easier for me

## Sample Config
```yaml
env:
  url: http://localhost:8080/api/v1
  vars:
    token: test
    id: 123213

requests:
  - name: "new token"
    method: "POST"
    endpoint: "/login"
    headers:
      "Content-type": "application/json"
    body: >
      {
        "username": "test",
        "password": "secret"
      }
    wantStatus: 200
    wantResponse: "{access_token}"
    vars:
      token: "access_token"
  - name: "get users"
    method: "GET"
    endpoint: "/users"
    headers:
      "Content-type": "application/json"
      "Authorization": "Bearer {{ .token }}"
    wantStatus: 200
    wantResponse: "{success}"
    vars:
      id: "data[0],id"
```

Vars can be used as contasts defined at the head of the config or assigned at run time with the data of the requests.
These vars can be used in the endpoint and headers with the ```{{ .variable }}``` notation.

The key ```vars``` inside the definition of the request will search for that key inside the response and assign it to the variable.
```yaml
env:
  ###
  vars:
    token: test

requests:
  - name: "new token"
    method: "POST"
    endpoint: "/login"
    headers:
      "Content-type": "application/json"
    body: >
      {
        "username": "test",
        "password": "secret"
      }
    wantStatus: 200
    wantResponse: "{access_token}"
    vars:
      token: "access_token" # <-- assign value to the variable token
```
The definition of token inside vars will search for a key ```"access_token"``` inside the response and assign the variable ```token``` its value.
If the variable is never defined inside a request ```vars``` key, it will retain its default value.

More nested is the data more keys divided by comma will translate in that deeper depth as well (it is always left first).
```yaml
env:
  ###
  vars:
    token: test

requests:
  - name: "get users"
    method: "GET"
    endpoint: "/users"
    headers:
      "Content-type": "application/json"
      "Authorization": "Bearer {{ .token }}"
    wantStatus: 200
    wantResponse: "{success}"
    vars:
      id: "data[0],id" # <-- assign value
```

In this scenario we want to assign a value dynamically to the ```id``` variable from the ```/users``` endpoint response.
This is the ```/users``` output:
```json
{
  "data": [
    {
      "age": 13,
      "id": "63995ba6-1d89-XXXX-9703-880a9cbdb986",
      "name": "Name Surname",
      "status": "",
      "username": "test"
    }
  ],
  "status": "success"
}
```
```data``` is the first key which happens to be an array, so we need to define which **i-th** of it as well like any normal array. ```id``` is the key of said array.

This would mean the variable *id* from its original value of **123213** is now **63995ba6-1d89-XXXX-9703-880a9cbdb986**.

## Build
```
make build
```
Bin is located in ```./cmd/build/courier```

## Usage

```
courier -f myconfig.yaml
courier # defaults to using courier.yaml or .courier/config.yaml in your current pwd
```

The output can be piped with ```jq```
```
courier -f myconfig.yaml | jq
courier | jq
```

The flag ```-o``` to open each response inside the editor of preference, uses the enviroment $EDITOR value.
```
courier -o
courier -o -json # json flag to have it prettier
courier -o -e nano # e flag to change the editor at need
```

The flag ```-t``` runs a test on all requests checking if the response status code matches with ```wantStatus``` and if the response contains ```wantResponse``` string.
```
courier -t
```
