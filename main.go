package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// Util functions block start
func readFile(filename string) []byte {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

func createFileIfNotExists(filename string) {
	f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
}

func MarshalUsers(users []User) []byte {
	out, err := json.Marshal(users)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func MarshalUser(user User) []byte {
	out, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func UnmarshalUsers(bytes []byte) []User {
	var out []User
	err := json.Unmarshal(bytes, &out)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func UnmarshalUser(user string) User {
	var out User
	err := json.Unmarshal([]byte(user), &out)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

// Util functions block end

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func list(filename string) []byte {
	createFileIfNotExists(filename)
	bytes := readFile(filename)
	return bytes
}

func findById(filename string, id string) []byte {
	createFileIfNotExists(filename)

	bytes := readFile(filename)
	if len(bytes) == 0 {
		return []byte{}
	}
	users := UnmarshalUsers(bytes)
	for _, u := range users {
		if u.Id == id {
			return MarshalUser(u)
		}
	}
	return []byte{}
}

func add(filename string, item string) ([]byte, error) {
	createFileIfNotExists(filename)
	bytes := readFile(filename)
	if len(bytes) == 0 {
		return MarshalUsers([]User{UnmarshalUser(item)}), nil
	}
	users := UnmarshalUsers(bytes)
	user := UnmarshalUser(item)
	for _, u := range users {
		if u.Id == user.Id {
			return MarshalUsers(users), fmt.Errorf("Item with id %s already exists", user.Id)
		}
	}
	return MarshalUsers(append(users, user)), nil
}

func remove(filename string, id string) ([]byte, error) {
	createFileIfNotExists(filename)

	bytes := readFile(filename)
	if len(bytes) == 0 {
		return MarshalUsers([]User{}), fmt.Errorf("Item with id %s not found", id)
	}
	users := UnmarshalUsers(bytes)
	for i := 0; i < len(users); i++ {
		if users[i].Id == id {
			return MarshalUsers(append(users[:i], users[i+1:]...)), nil
		}
	}
	return MarshalUsers(users), fmt.Errorf("Item with id %s not found", id)
}

type Arguments map[string]string

func Perform(args Arguments, writer io.Writer) error {

	operation, id, item, fileName := args["operation"], args["id"], args["item"], args["fileName"]

	if len(operation) == 0 {
		return errors.New("-operation flag has to be specified")
	}
	switch operation {
	case "list":
		if len(fileName) == 0 {
			return errors.New("-fileName flag has to be specified")
		}
		listsing := list(fileName)
		if len(listsing) != 0 {
			writer.Write(listsing)
		}
	case "add":
		if len(fileName) == 0 {
			return errors.New("-fileName flag has to be specified")
		}
		if len(item) == 0 {
			return errors.New("-item flag has to be specified")
		}
		users, err := add(fileName, item)
		if err != nil {
			writer.Write([]byte(err.Error()))
		}
		os.WriteFile(fileName, users, 0644)
	case "findById":
		if len(fileName) == 0 {
			return errors.New("-fileName flag has to be specified")
		}
		if len(id) == 0 {
			return errors.New("-id flag has to be specified")
		}
		bytes := findById(fileName, id)
		writer.Write(bytes)
	case "remove":
		if len(fileName) == 0 {
			return errors.New("-fileName flag has to be specified")
		}
		if len(id) == 0 {
			return errors.New("-id flag has to be specified")
		}
		bytes, err := remove(fileName, id)
		if err != nil {
			writer.Write([]byte(err.Error()))
		} else {
			remove(fileName, id)
			os.WriteFile(fileName, bytes, 0644)
			writer.Write(bytes)
		}
	default:
		return errors.New("Operation abcd not allowed!")
	}

	return nil
}

func parseArgs() Arguments {
	operation := flag.String("operation", "", "operation")
	filaName := flag.String("fileName", "user.json", "file name")
	item := flag.String("item", "", "item")
	id := flag.String("id", "", "id")

	flag.Parse()

	arguments := Arguments{
		"operation": *operation,
		"fileName":  *filaName,
		"item":      *item,
		"id":        *id,
	}

	return arguments
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
