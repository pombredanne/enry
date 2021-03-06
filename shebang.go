package enry

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

const shebang = `#!`

var (
	shebangExecHack = regexp.MustCompile(`exec (\w+).+\$0.+\$@`)
	pythonVersion   = regexp.MustCompile(`python\d\.\d+`)
)

// GetLanguagesByShebang returns a slice of possible languages for the given content, filename will be ignored.
// It accomplish the signature to be a Strategy type.
func GetLanguagesByShebang(filename string, content []byte) (languages []string) {
	interpreter := getInterpreter(content)
	return languagesByInterpreter[interpreter]
}

func getInterpreter(data []byte) (interpreter string) {
	line := getFirstLine(data)
	if !hasShebang(line) {
		return ""
	}

	// skip shebang
	line = bytes.TrimSpace(line[2:])

	splitted := bytes.Fields(line)
	if bytes.Contains(splitted[0], []byte("env")) {
		if len(splitted) > 1 {
			interpreter = string(splitted[1])
		}
	} else {

		splittedPath := bytes.Split(splitted[0], []byte{'/'})
		interpreter = string(splittedPath[len(splittedPath)-1])
	}

	if interpreter == "sh" {
		interpreter = lookForMultilineExec(data)
	}

	if pythonVersion.MatchString(interpreter) {
		interpreter = interpreter[:strings.Index(interpreter, `.`)]
	}

	return
}

func getFirstLine(data []byte) []byte {
	buf := bufio.NewScanner(bytes.NewReader(data))
	buf.Scan()
	line := buf.Bytes()
	if err := buf.Err(); err != nil {
		return nil
	}

	return line
}

func hasShebang(line []byte) bool {
	shebang := []byte(shebang)
	return bytes.HasPrefix(line, shebang)
}

func lookForMultilineExec(data []byte) string {
	const magicNumOfLines = 5
	interpreter := "sh"

	buf := bufio.NewScanner(bytes.NewReader(data))
	for i := 0; i < magicNumOfLines && buf.Scan(); i++ {
		line := buf.Bytes()
		if shebangExecHack.Match(line) {
			interpreter = shebangExecHack.FindStringSubmatch(string(line))[1]
			break
		}
	}

	if err := buf.Err(); err != nil {
		return interpreter
	}

	return interpreter
}
