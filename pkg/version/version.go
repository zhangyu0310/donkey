package version

import (
	"fmt"
	"strings"
)

func init() {
	parseVersionInfo()
}

// Version variables should be embedded by ldflags
var (
	BuildTS       = "?" // `date '+%Y%m%d-%H%M'`
	CompileInfo   = "?" // `uname -p`
	GitBranch     = "?" // `git symbolic-ref --short -q HEAD`
	GitCommitDate = "?" // `git log -1 --pretty=format:%ci`
	GitHash       = "?" // `git rev-parse HEAD`
	GOVersion     = "?" // `go version`
	MajorVersion  = "?" // [0].0.1-snapshot
	MinorVersion  = "?" // 0.[0].1-snapshot
	PatchVersion  = "?" // 0.0.[1]-snapshot
	SuffixVersion = ""  // 0.0.1-[snapshot]

	BaseVersionInfo     = ""
	CompilerVersionInfo = ""
	CompileTime         = ""
	FullVersionInfo     = ""
)

// parseVersionInfo parse and arrange version variables
func parseVersionInfo() {
	MajorVersion = strings.TrimSpace(MajorVersion)
	MinorVersion = strings.TrimSpace(MinorVersion)
	PatchVersion = strings.TrimSpace(PatchVersion)
	SuffixVersion = strings.TrimSpace(SuffixVersion)

	if strings.Contains(GOVersion, "version") {
		CompilerVersionInfo = GOVersion[strings.Index(GOVersion, "version")+8:]
	}

	BaseVersionInfo = fmt.Sprintf("%s.%s.%s", MajorVersion, MinorVersion, PatchVersion)
	FullVersionInfo = BaseVersionInfo
	CompileTime = BuildTS
	if len(strings.TrimSpace(SuffixVersion)) > 0 {
		FullVersionInfo = fmt.Sprintf("%s-%s", FullVersionInfo, SuffixVersion)
	}
}

// VerInfo results in format:
//
// Version        : 0.0.1-dirty
// Compile Commit : aba0ea67e235a52cb1a76cdc1e79ab615be49a4b
// Compile Branch : master
// Git Commit Date: 2022-09-30 17:54:21 +0800
// Compile Time   : 20231031-1656
// CPU            : arm
// Compiler       : go1.21.0 darwin/arm64
func VerInfo() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintln("Version        :", FullVersionInfo))
	s.WriteString(fmt.Sprintln("Compile Commit :", GitHash))
	s.WriteString(fmt.Sprintln("Compile Branch :", GitBranch))
	s.WriteString(fmt.Sprintln("Git Commit Date:", GitCommitDate))
	s.WriteString(fmt.Sprintln("Compile Time   :", CompileTime))
	s.WriteString(fmt.Sprintln("CPU            :", CompileInfo))
	s.WriteString(fmt.Sprintln("Compiler       :", CompilerVersionInfo))
	return s.String()
}
