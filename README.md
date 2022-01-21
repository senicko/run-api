# run-api

Run api allows to execute code inside docker containers.

## Usage

`POST /run`

### Desc

Executes the code inside docker container.

### Request

Requets body is a json configuration object passed to [bee](https://github.com/senicko/bee). Checkout bee's readme for more info.

- language `target language`
- files `array of input files`
  - name `name of the file`
  - content `content of the file`
