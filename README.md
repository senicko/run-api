# run-api

Run api allows to execute code inside docker containers.

## Usage

`POST /run`

### Desc

Executes the code inside docker container.

### Request

Example request that prints text passed with stdin.

```json
{
    "config": {
        "language": "golang"
    },
    "stdin": "test",
    "files": [
        {
            "name": "main.go",
            "body": "package main \n import (\n \"fmt\" \n \"bufio\" \n \"os\" \n ) \n func main() { reader := bufio.NewReader(os.Stdin) \n name, _ := reader.ReadString('\\n') \n fmt.Println(name) }"
        }
    ]
}
```
