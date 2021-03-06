package lmd

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// ReadAFile is a convenience function. Given a filename string, reads the file and passes it back to
// the calling function as a string. Given a "-" switch the function will read from stdin rather than
// from a file.
func ReadAFile(file_to_read string) string {

	if file_to_read == " -" || file_to_read == "-" {

		std_in_read, std_in_err := ioutil.ReadAll(os.Stdin)
		if std_in_err != nil {
			log.Fatal(std_in_err)
		}
		return string(std_in_read)
	}

	file_buffer, file_read_err := ioutil.ReadFile(file_to_read)

	if file_read_err != nil {
		log.Fatal(file_read_err)
	}

	contents := string(file_buffer)
	return contents
}

// importIncludedFiles handles importing files into the primary contents string. First it compiles a
// regular expression which will search for the trigger string `@include PARTIAL`.
//
// If one or more match is found, the function will simply replace the `@include PARTIAL` line with the
// read in string of the included partial. The complete string will be returned to the calling function.
func importIncludedFiles(fileContents string) string {
	importRegExp := regexp.MustCompile(`(?m)^@include (.*?)$`)

	if importRegExp.MatchString(fileContents) {
		importedFiles := importRegExp.FindAllStringSubmatch(fileContents, -1)
		for _, importedFile := range importedFiles {
			fileContents = strings.Replace(fileContents, importedFile[0], ReadAFile(importedFile[1]), -1)
		}
		return fileContents
	} else {
		return fileContents
	}
}

// parseTemplateToFindParameters handles paramaters which are passed to the parser either separately from the
// template file or as part of the template file. This function manages the process of stripping paramaters
// out of a template file. The function first compiles a YAML Front Matter regular expression. Then if a
// match for that regular expression is found, the contents of the template file are replaced with an empty
// string and the YAML front matter is returned, along with the replaced contents (both as strings) to the
// calling function.
func parseTemplateToFindParameters(fileContents string) (string, string) {
	yamlRegExp := regexp.MustCompile(`(?sm)\A(---\s*\n.*?)(^---\s*\n)`)

	if yamlRegExp.MatchString(fileContents) {
		yamlFrontMatter := yamlRegExp.FindAllStringSubmatch(fileContents, -1)[0][1]
		fileContents = yamlRegExp.ReplaceAllString(fileContents, "")
		return yamlFrontMatter, fileContents
	} else {
		return "", fileContents
	}
}

// unmarshallParameters unmarshalls paramaters either in yaml (TBD) or json into the paramaters map. This
// function is responsible for unmarshalling the paramaters from yaml or json strings into (first a byte
// array) and subsequently into the paramaters map which is returned to the calling function.
func unmarshallParameters(parameters string) map[string]string {
	parameter_bytes := []byte(parameters)
	param := make(map[string]string)
	yaml.Unmarshal(parameter_bytes, &param)
	for key, val := range param {
		if strings.TrimSpace(val) == "@today" {
			year, month, day := time.Now().Date()
			date := fmt.Sprintf("%v %v %v", day, month, year)
			param[key] = date
		} else if strings.TrimSpace(val) == "@today_us" {
			year, month, day := time.Now().Date()
			date := fmt.Sprintf("%v %v, %v", month, day, year)
			param[key] = date
		}
	}
	return param
}

// mergeParameters is a convenience function which will merge two hash maps into one. Any conflicting parameters
// in the two maps will be resolved in favor of the *first* map which is passed. That is to say that the first
// map passed to the function, the `superior_map` map, will overwrite the `sublimated_map`.
func mergeParameters(superior_map map[string]string, sublimated_map map[string]string) map[string]string {
	for k, v := range superior_map {
		sublimated_map[k] = v
	}
	return sublimated_map
}

// jsonizeParameters is a convenience function which will simply marshal the parameters map and return
// that marshaled json as a string to the calling function.
func jsonizeParameters(parameters map[string]string) string {

	paramsAsJsonByteArray, err := json.Marshal(parameters)

	if err != nil {
		log.Fatal("Could not read the parameters")
	}

	return string(paramsAsJsonByteArray)

}
