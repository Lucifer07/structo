package main

import (
	"fmt"

	structo "github.com/Lucifer07/Structo"
)

type Sample struct {
	Name   string  `json:"name"`
	Sample Sample2 `json:"sampleData"`
}
type Sample2 struct {
	Number int `json:"number"`
}

type Sample3 struct {
	Name   string  `json:"name"`
	Sample Sample4 `json:"sampleData"`
}
type Sample4 struct {
	Number int `json:"number"`
}

func main() {
	structor := structo.NewStructo()

	sampleData := Sample{
		Name: "test",
		Sample: Sample2{
			Number: 1,
		},
	}

	fmt.Println("data sample ", sampleData)
	//  example to encode.
	encoded, err := structor.EncodeToString(sampleData)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("result encrypt :", encoded)

	//  example to decode.
	var dest Sample

	err = structor.DecodeFromString(encoded, &dest)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("result decode :", dest)

	// example to encode safe.
	encoded, err = structor.EncodeToStringSafe(sampleData)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("result encode :", encoded)

	// example to decode safe
	var destiny Sample

	err = structor.DecodeToStringSafe(encoded, &destiny)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("result decode", destiny)

	// example copy

	var destinyCopy Sample3

	structo.Copy(&destinyCopy, sampleData)

	fmt.Println("sampleData : ", sampleData)
	fmt.Println("destinyCopy : ", destinyCopy)
}
