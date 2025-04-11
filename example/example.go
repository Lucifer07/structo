package main

import (
	"fmt"

	structo "github.com/Lucifer07/Structo"
)

type Address struct {
	City    string `json:"city"`
	ZipCode string `json:"zipCode"`
}

type User struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Email    *string  `json:"email"`
	Tags     []string `json:"tags"`
	Metadata Metadata `json:"metadata"`
	Address  *Address `json:"address"`
}

type Metadata struct {
	Active  bool   `json:"active"`
	Version string `json:"version"`
}

type ExternalUser struct {
	ID      int
	Name    string
	Address Address
	Active  bool
}

type Profile struct {
	Name    string   `default:"Anonymous"`
	Age     int      `default:"18"`
	Active  bool     `default:"true"`
	Tag     []uint `default:"1,2,3"`
	Email   *string  `default:"default@example.com"`
	Address ProfileAddress
}

type ProfileAddress struct {
	City    string `default:"Unknown City"`
	ZipCode string `default:"00000"`
}

func main() {
	user := createSampleUser()

	structor := structo.NewConverter()

	// Encode & Decode
	encodedStr := encodeToString(structor, user)
	decodeFromString(structor, encodedStr)

	// Binary Encode & Decode
	binaryData := structToBinary(structor, user)
	binaryToStruct(structor, binaryData)

	// Safe Encode & Decode
	safeEncoded := encodeToStringSafe(structor, user)
	decodeFromStringSafe(structor, safeEncoded)

	// Flatten & Unflatten
	flattened := flatten(user)
	unflatten(flattened)

	// Diff & Track
	updatedUser := user
	updatedUser.Name = "John Doe"
	updatedUser.Metadata.Active = false
	updatedUser.Tags = append(updatedUser.Tags, "backend")

	printDiff(user, updatedUser)

	printTrack(user, updatedUser)

	// Copy to Different Struct Shape
	var extUser ExternalUser
	structo.Copy(&extUser, user)
	fmt.Println("\nCopied Struct (to different shape):", extUser)

	printProfile()
}

func createSampleUser() User {
	email := "user@example.com"
	return User{
		ID:    101,
		Name:  "Jane Doe",
		Email: &email,
		Tags:  []string{"golang", "developer"},
		Metadata: Metadata{
			Active:  true,
			Version: "v1.2.3",
		},
		Address: &Address{
			City:    "Jakarta",
			ZipCode: "12345",
		},
	}
}

func encodeToString(conv structo.Converter, data any) string {
	encoded, err := conv.EncodeToString(data)
	check(err, "encode string")
	fmt.Println("Encoded:", encoded)
	return encoded
}

func decodeFromString(conv structo.Converter, encoded string) {
	var user User
	err := conv.DecodeFromString(encoded, &user)
	check(err, "decode string")
	fmt.Printf("Decoded: %+v\n \n", user)
}

func structToBinary(conv structo.Converter, data any) []byte {
	bin, err := conv.StructToBinary(data)
	check(err, "struct to binary")
	fmt.Println("Binary:", bin)
	return bin
}

func binaryToStruct(conv structo.Converter, bin []byte) {
	var user User
	err := conv.BinaryToStruct(bin, &user) // Use DecodeFromBinary instead
	check(err, "binary to struct")
	fmt.Printf("Struct:  %+v\n \n", user)
}

func encodeToStringSafe(conv structo.Converter, data any) string {
	encoded, err := conv.EncodeToStringSafe(data)
	check(err, "encode safe")
	fmt.Println("Encoded Safe:", encoded)
	return encoded
}

func decodeFromStringSafe(conv structo.Converter, encoded string) {
	var user User
	err := conv.DecodeFromStringSafe(encoded, &user)
	check(err, "decode safe")
	fmt.Printf("Decoded Struct Safe:  %+v\n \n", user)
}

func flatten(user User) map[string]interface{} {
	flat := structo.Flatten(user)
	fmt.Println("\nFlattened Struct:")
	for k, v := range flat {
		fmt.Printf("  %s = %v\n", k, v)
	}
	return flat
}
func unflatten(flat map[string]any) {
	var user User
	err := structo.Unflatten(flat, &user)
	check(err, "unflatten")
	fmt.Printf("\nUnflattened Struct:  %+v\n \n", user)
}

func printDiff(oldUser, newUser User) {
	diff, err := structo.Diff(oldUser, newUser)
	check(err, "diff")
	fmt.Println("\nDiff:")
	for k, v := range diff {
		fmt.Printf("  %s: %v -> %v\n", k, v[0], v[1])
	}
}

func printTrack(oldUser, newUser User) {
	changes, err := structo.TrackWithHistory(oldUser, newUser)
	check(err, "track")
	fmt.Println("\nChanges:")
	for field, res := range changes {
		switch res.Action {
		case structo.Add:
			fmt.Printf("  %s: appended %v\n", field, res.Data.GetData())
		case structo.Remove:
			fmt.Printf("  %s: removed %v\n", field, res.Data.GetData())
		case structo.Change:
			data := res.Data.GetData().([]any)
			fmt.Printf("  %s: changed from %v to %v\n", field, data[0], data[1])
		}
	}
}

func printProfile() {
	var p Profile
	err := structo.InjectDefaults(&p)
	check(err, "inject defaults")
	fmt.Printf("%+v\n", p)
}

func check(err error, context string) {
	if err != nil {
		panic(fmt.Sprintf("[%s] %v", context, err))
	}
}
