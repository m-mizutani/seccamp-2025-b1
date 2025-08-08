package logcore

import (
	"fmt"
	"math/rand"
	"strings"
)

func ShowSampleResources() {
	resources := generateResources()
	
	fmt.Println("Sample resources generated:")
	fmt.Println(strings.Repeat("=", 80))
	
	// ランダムに30個のリソースを表示
	rand.Shuffle(len(resources), func(i, j int) {
		resources[i], resources[j] = resources[j], resources[i]
	})
	
	for i := 0; i < 30 && i < len(resources); i++ {
		r := resources[i]
		fmt.Printf("- %s (Type: %s, Visibility: %s)\n", r.Name, r.Type, r.Visibility)
	}
}

func main() {
	ShowSampleResources()
}