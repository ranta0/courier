# courier

Http client to make testing apis, or just making normal curl requests, a bit
easier for me. Automation and whatnot.

## Config

```yaml
vars:
  url: http://localhost:8080/api/v1
  token: test
  id: 123456

requests:
  - name: "new token"
    method: "POST"
    endpoint: "{{ .url }}/login"
    headers:
      "Content-type": "application/json"
    body: >
      {
        "username": "test",
        "password": "secret"
      }
    wantStatus: 200
    wantResponse: "{access_token}"
    delay: 1
    vars:
      token: "access_token"
  - name: "get users"
    method: "GET"
    endpoint: "{{ .url }}/users"
    headers:
      "Content-type": "application/json"
      "Authorization": "Bearer {{ .token }}"
    wantStatus: 200
    wantResponse: "{success}"
    vars:
      id: "data[0].id"
```

The key `delay` delays the current request by X seconds, X must be an integer.

Variables are defined in the `vars` key, they can be used as they are or reassigned.
These variables can be used in the `endpoint`, `headers`, `body` and `name`
with the `{{ .var_name }}` notation.

The key `vars` inside the definition of the `request` will search for that key
inside the response and assign it to that variable if found.

```yaml
env:
  ###
  vars:
    token: test

requests:
  - name: "new token"
    method: "POST"
    endpoint: "{{ .url }}/login"
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

For instance, this will attempt to set token to a key `access_token` taken from
the request reponse output.

```yaml
vars:
  token: "access_token"
```

If the variable is never defined inside a request `vars` key, it will retain
its default value.

To access nested values inside a response we use a `dot` notation. Treating the
output as an object. The more the dots, the deeper the depth inside the object

```yaml
env:
  ###
  vars:
    token: test

requests:
  - name: "get users"
    method: "GET"
    endpoint: "{{ .url }}/users"
    headers:
      "Content-type": "application/json"
      "Authorization": "Bearer {{ .token }}"
    wantStatus: 200
    wantResponse: "{success}"
    vars:
      id: "data[0].id" # <-- assign value
```

In this scenario we want to assign a value dynamically to the `id` variable from the `/users` endpoint response.
Given this as the `/users` output:

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

`data` is the first key which happens to be an array, so we need to define
which **i-th** of it as well like any normal array. `id` is the key of said array.

This would mean the variable _id_ from its original value of **123213** is
now **63995ba6-1d89-XXXX-9703-880a9cbdb986**.

## Build

```
make build
```

Bin is located in `cmd/build`

## Usage

```
courier -config-file myconfig.yaml
courier # defaults to using courier.yaml or .courier/config.yaml in your current pwd
```

The output can be piped with `jq`

```
courier -config-file myconfig.yaml | jq
courier | jq
```

The flag `-open` to open each response inside the editor of preference, uses the enviroment $EDITOR value.

```
courier -open
courier -open -json # json flag to have it prettier
courier -open -editor nano # e flag to change the editor at need
```

The flag `-test` runs a test on all requests checking if the response status code matches with `wantStatus` and if the response contains `wantResponse` string.

```
courier -test
```
