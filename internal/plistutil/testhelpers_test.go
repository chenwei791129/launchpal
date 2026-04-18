package plistutil

import (
	"fmt"
	"os/exec"
)

// convertXMLToBinaryForTest converts an XML plist file to binary format using plutil,
// producing a file suitable for round-trip testing of NormalizeFromPath.
func convertXMLToBinaryForTest(xmlPath, binaryPath string) error {
	cmd := exec.Command("plutil", "-convert", "binary1", "-o", binaryPath, xmlPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("plutil convert: %w (output: %s)", err, string(out))
	}
	return nil
}
