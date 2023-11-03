package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/glair-ai/glair-vision-go"
	"github.com/glair-ai/glair-vision-go/client"
)

func main() {
	ctx := context.Background()

	config := glair.NewConfig("sdk-tester", "bzJ0Vt0a8R2XqVbCPrgH", "OKMjceKWLXdmTKYwoSCXAVVDWtQWRrhr")
	client := client.New(config)

	file, _ := os.Open("../images/ktp.jpeg")

	result, err := client.Ocr.Ktp(ctx, file)

	if err != nil {
		log.Fatalln(err.Error())
	}

	beautified, _ := json.MarshalIndent(result, "", "  ")

	fmt.Println(string(beautified))
}