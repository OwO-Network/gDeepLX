# gDeepLX
gDeepLX is a Go library used for DeepL translation.

## Deprecated

Core repository [OwO-Network/DeepLX](https://github.com/OwO-Network/DeepLX) already supports more features.

## Installation

Install it with the go get command:
```bash
go get github.com/OwO-Network/gdeeplx
```

## Usage
Then, you can create a new DeepL translation client and use it for translation:

```go
import (
	"fmt"
	"github.com/OwO-Network/gdeeplx"
)

func main() {
	result, err := gdeeplx.Translate("Hello World!", "EN", "ZH", 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(result)
}
```
## Author

**gDeepLX** © [Vincent Young](https://github.com/missuo), Released under the [MIT](./LICENSE) License.<br>
