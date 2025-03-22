# 🚀 Structo

Structo is a lightweight and efficient Go package for serializing and deserializing structs into binary and string formats. It also provides encryption and deep copying capabilities.

## 📦 Installation

Install Structo using `go get`:

```sh
go get github.com/Lucifer07/structo
```

## 🌟 Features

- Convert structs to binary format and vice versa.
- Encode and decode to/from strings.
- Secure encoding with encryption.
- Deep copy struct data.

## 🚀 Usage

### Import Structo

```go
package main

import (
	"fmt"

	structo "github.com/Lucifer07/structo"
)
```

### Define Structs

```go
type Sample struct {
	Name   string  `json:"name"`
	Sample Sample2 `json:"sampleData"`
}

type Sample2 struct {
	Number int `json:"number"`
}
```

### Encode and Decode Example

```go
func main() {
	structor := structo.NewStructo()

	sampleData := Sample{
		Name: "test",
		Sample: Sample2{
			Number: 1,
		},
	}

	fmt.Println("Original data:", sampleData)

	// Encoding
	encoded, err := structor.EncodeToString(sampleData)
	if err != nil {
		fmt.Println("Encoding error:", err)
		return
	}
	fmt.Println("Encoded:", encoded)

	// Decoding
	var decoded Sample
	err = structor.DecodeFromString(encoded, &decoded)
	if err != nil {
		fmt.Println("Decoding error:", err)
		return
	}
	fmt.Println("Decoded:", decoded)
}
```

### Secure Encoding & Decoding Example

```go
// Secure Encoding
encodedSafe, err := structor.EncodeToStringSafe(sampleData)
if err != nil {
	fmt.Println("Encoding safe error:", err)
	return
}
fmt.Println("Encoded Safe:", encodedSafe)

// Secure Decoding
var decodedSafe Sample
err = structor.DecodeToStringSafe(encodedSafe, &decodedSafe)
if err != nil {
	fmt.Println("Decoding safe error:", err)
	return
}
fmt.Println("Decoded Safe:", decodedSafe)
```

### Struct Deep Copy Example

```go
type Sample3 struct {
	Name   string  `json:"name"`
	Sample Sample4 `json:"sampleData"`
}

type Sample4 struct {
	Number int `json:"number"`
}

var copied Sample3
structo.Copy(&copied, sampleData)
fmt.Println("Copied Struct:", copied)
```

## 📜 License

Structo is licensed under the MIT License.

## 💡 Contributions

Contributions are welcome! Feel free to open issues and pull requests to improve Structo.

---

🔥 **Enjoy fast and secure struct serialization with Structo!**

