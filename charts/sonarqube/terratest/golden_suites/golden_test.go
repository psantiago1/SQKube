package golden

import (
	"flag"
	"io/ioutil"
	"regexp"
    "path/filepath"
    "strings"
    "testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/suite"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update-golden", false, "update golden test output files")
var runOneTest = flag.String("runOneTest","","allow to run one specific test from the suite")

type TemplateGoldenTest struct {
	suite.Suite
	ChartPath      string
	SuiteName      string
	Release        string
	Namespace      string
	GoldenFileName string
	Templates      []string
	IgnoredLines   []string
}

func getGoldenFileNames(relativeCasePath string) []string {

    goldenFileNames := []string{}

    items, _ := ioutil.ReadDir(relativeCasePath)
    for _, item := range items {
        if ! item.IsDir() {
            r, _ := regexp.Compile("(.*)\\.golden\\.yaml")
            if r.MatchString(item.Name()) {
            captureResult := r.FindStringSubmatch(item.Name())
            goldenFileNames = append(goldenFileNames, captureResult[1])
            }
        }
    }

    return goldenFileNames
}

func getCaseDirectoriesNames() []string {

    caseDirectoriesNames := []string{}

    items, _ := ioutil.ReadDir(".")
    for _, item := range items {
        if item.IsDir() {
            r, _ := regexp.Compile(".*_case$")
            if r.MatchString(item.Name()) {
            caseDirectoriesNames = append(caseDirectoriesNames, item.Name())
            }
        }
    }

    return caseDirectoriesNames
}

func (s *TemplateGoldenTest) TestGolden() {
    if *runOneTest != "" && *runOneTest != s.SuiteName {
    s.T().Skip("runOneTest is passed, skiping this one")
    }

    suiteAbsPath, err := filepath.Abs(s.SuiteName+"/values.yaml")

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", s.Namespace),
		ValuesFiles: []string{suiteAbsPath},
	}
	goldenFile := s.SuiteName + "/" + s.GoldenFileName + ".golden.yaml"

	output := helm.RenderTemplate(s.T(), options, s.ChartPath, s.Release, s.Templates)

    s.IgnoredLines = append(s.IgnoredLines, `(?m)^#.*`)// remove comment we add in our templates
    s.IgnoredLines = append(s.IgnoredLines, `(?m)\schecksum/.*`)// remove comment we add in our templates
    s.IgnoredLines = append(s.IgnoredLines, `(?m)\sapp.kubernetes.io/version.*`)// remove that does nos structurally matters
    s.IgnoredLines = append(s.IgnoredLines, `(?m)\schart:.*`)// remove comment we add in our templates
	s.IgnoredLines = append(s.IgnoredLines, `\n$`)//prevent last empty line from being an issue between IDE save and our helm templates
	s.IgnoredLines = append(s.IgnoredLines, `(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)// this is to remove all empty lines we misleadly generate with our templates.

	bytes := []byte(output)
	expected, err := ioutil.ReadFile(goldenFile)

	for _, ignoredLine := range s.IgnoredLines {
		regex := regexp.MustCompile(ignoredLine)
		bytes = regex.ReplaceAll(bytes, []byte(""))
		expected = regex.ReplaceAll(expected, []byte(""))
	}
	output = string(bytes)

	if *update {
		err := ioutil.WriteFile(goldenFile, bytes, 0644)
		s.Require().NoError(err, "Golden file was not writable")
	}

	// then
	s.Require().NoError(err, "Golden file doesn't exist or was not readable")
	s.Require().Equal(string(expected), output)
}

func TestGolden(t *testing.T) {
	t.Parallel()

    chartPath, err := filepath.Abs("../../")
    require.NoError(t, err)

	caseDirectoriesNames := getCaseDirectoriesNames()

	for _, caseDirectoriesName := range caseDirectoriesNames {

        goldenFileNames := getGoldenFileNames(caseDirectoriesName)
        for _,goldenFileName := range goldenFileNames{
            suite.Run(t, &TemplateGoldenTest{
                ChartPath:      chartPath,
                SuiteName:      caseDirectoriesName,
                Release:        "release-name",
                Namespace:      "sonarqube-helm-" + strings.ToLower(random.UniqueId()),
                GoldenFileName: goldenFileName,
                Templates:      []string{"templates/" + goldenFileName + ".yaml"},
            })
        }
	}

}
