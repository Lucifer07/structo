# Structo 🏗️  
Go library for **struct serialization**, **diffing**, **change tracking**, and **default injection** with nested struct, pointer, slice, and encoding support.

---

## ✨ Features

- ✅ Encode/decode struct to string or binary
- 🔄 Flatten dan unflatten nested struct
- 📆 Inject default values to struct
- 🕵️ Struct diffing and history tracking
- ⟳ Copy between structs with different shapes

---

## 📦 Installations

```bash
go get github.com/Lucifer07/Structo
```

---

## 🚀 Example

### Struct Sample

```go
type Address struct {
	City    string `json:"city"`
	ZipCode string `json:"zipCode"`
}

type Metadata struct {
	Active  bool   `json:"active"`
	Version string `json:"version"`
}

type User struct {
	ID       int
	Name     string
	Email    *string
	Tags     []string
	Metadata Metadata
	Address  *Address
}
```

---

### 🔐 Encode & Decode

```go
conv := structo.NewConverter()
encoded := conv.EncodeToString(user)
conv.DecodeFromString(encoded, &user)
```

---

### 🧱 Binary Encode & Decode

```go
bin := conv.StructToBinary(user)
conv.BinaryToStruct(bin, &user)
```

---

### 🔐 Safe Encode 

```go
encoded := conv.EncodeToStringSafe(user)
conv.DecodeFromStringSafe(encoded, &user)
```

---

### 🧬 Flatten & Unflatten

```go
flat := structo.Flatten(user)
// map[string]any: {"Name":"Jane", "Address.City":"Jakarta", ...}

var u User
structo.Unflatten(flat, &u)
```

---

### 🗓️ Diff Struct

```go
diff, _ := structo.Diff(oldUser, newUser)
for field, values := range diff {
	fmt.Printf("%s: %v -> %v\n", field, values[0], values[1])
}
```

---

### 📊 Track Changes (Add, Remove, Change)

```go
changes, _ := structo.TrackWithHistory(oldUser, newUser)
for field, res := range changes {
	switch res.Action {
	case structo.Add:
		fmt.Printf("%s: appended %v\n", field, res.Data.GetData())
	case structo.Remove:
		fmt.Printf("%s: removed %v\n", field, res.Data.GetData())
	case structo.Change:
		data := res.Data.GetData().([]any)
		fmt.Printf("%s: changed from %v to %v\n", field, data[0], data[1])
	}
}
```

---

### ⟳ Copy Struct to a Different Shape

```go
var external ExternalUser
structo.Copy(&external, user)
```

---

### ⚙️ Inject Default Values

```go
type Profile struct {
	Name  string `default:"Anonymous"`
	Age   int    `default:"18"`
	Email string `default:"default@example.com"`
}
var p Profile
structo.InjectDefaults(&p)
// p.Name = "Anonymous", etc.
```

---

## 📁 Full Example

view the `example/main.go` file for full example, runnable example.

---

## 📜 Lisensi

MIT © Lucifer07

---

## 💡 Contributions

Contributions are welcome! Feel free to open issues and pull requests to improve Structo.

---

🔥 **Enjoy fast and secure struct serialization with Structo!**