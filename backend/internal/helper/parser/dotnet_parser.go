package parser

import (
	"regexp"
)

// DotNetParser handles parsing of .NET project files
type DotNetParser struct{}

// NewDotNetParser creates a new instance of DotNetParser
func NewDotNetParser() *DotNetParser {
	return &DotNetParser{}
}

// GetRuntime returns the runtime type for .NET
func (p *DotNetParser) GetRuntime() RuntimeType {
	return RuntimeDotNet
}

// Parse parses .csproj files
func (p *DotNetParser) Parse(content string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo

	// Parse PackageReference elements
	packageRefRegex := regexp.MustCompile(`<PackageReference\s+Include="([^"]+)"\s+Version="([^"]+)"`)
	matches := packageRefRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			depInfo := p.ParseDependency(match[1], match[2])
			dependencies = append(dependencies, *depInfo)
		}
	}

	return dependencies, nil
}

// ParseDependency parses a single .NET dependency
func (p *DotNetParser) ParseDependency(name, version string) *DependencyInfo {
	return &DependencyInfo{
		Name:    name,
		Owner:   "", // Would need NuGet API to determine
		Repo:    name,
		Version: version,
		Runtime: string(RuntimeDotNet),
	}
}

// GetRepositoryURL gets GitHub URL for .NET packages
func (p *DotNetParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Common .NET packages with known GitHub repositories
	commonPackages := map[string]string{
		"Microsoft.AspNetCore.App":      "https://github.com/dotnet/aspnetcore",
		"Microsoft.EntityFrameworkCore": "https://github.com/dotnet/efcore",
		"Newtonsoft.Json":               "https://github.com/JamesNK/Newtonsoft.Json",
		"AutoMapper":                    "https://github.com/AutoMapper/AutoMapper",
		"Serilog":                       "https://github.com/serilog/serilog",
		"NLog":                          "https://github.com/NLog/NLog",
		"xunit":                         "https://github.com/xunit/xunit",
		"NUnit":                         "https://github.com/nunit/nunit",
		"Moq":                           "https://github.com/moq/moq4",
		"FluentAssertions":              "https://github.com/fluentassertions/fluentassertions",
		"Microsoft.Extensions.DependencyInjection": "https://github.com/dotnet/extensions",
	}

	if url, exists := commonPackages[dep.Name]; exists {
		return url
	}

	// Try common patterns for Microsoft packages
	if dep.Repo != "" {
		return "https://github.com/" + dep.Repo + "/" + dep.Repo
	}

	return ""
}
