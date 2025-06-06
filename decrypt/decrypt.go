package decrypt

import (
	"bytes"
	_ "embed"
	"os"

	"github.com/djohts/tpc-truckersmp/utils"
)

//go:embed Decrypt.exe
var decrypt_bytes []byte

func EnsureDecrypt() *os.File {
	const decryptFile = "SII_Decrypt.exe"

	if utils.IsFile(decryptFile) {
		data, err := os.ReadFile(decryptFile)
		utils.HandleError(err)
		if bytes.Equal(data, decrypt_bytes) {
			f, err := os.Open(decryptFile)
			utils.HandleError(err)
			defer f.Close()
			return f
		}
		utils.HandleError(os.Remove(decryptFile))
	}
	return installDecrypt()
}

func installDecrypt() *os.File {
	err := os.WriteFile("SII_Decrypt.exe", decrypt_bytes, 0o755)
	utils.HandleError(err)

	f, err := os.Open("SII_Decrypt.exe")
	utils.HandleError(err)

	defer f.Close()

	return f
}
