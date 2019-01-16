package printer

import (
	"encoding/json"
	"fmt"
	"os"
)

func JSONOutput(v interface{}) {
	out, err := json.MarshalIndent(v, "", "    ")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(string(out))
}
