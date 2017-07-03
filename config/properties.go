package config

import (
	"bytes"
	"io"

	"os"
	"path/filepath"

	"fmt"

	"github.com/getgauge/common"
)

const comment = `This file contains Gauge specific internal configurations. Do not delete`

type property struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	description  string
	defaultValue string
}

type properties struct {
	p map[string]*property
}

func (p *properties) set(k, v string) error {
	if _, ok := p.p[k]; ok {
		p.p[k].Value = v
		return nil
	}
	return fmt.Errorf("Config '%s' doesn't exist.", k)
}

func (p *properties) get(k string) (*property, error) {
	if _, ok := p.p[k]; ok {
		return p.p[k], nil
	}
	return nil, fmt.Errorf("Config '%s' doesn't exist.", k)
}

func (p *properties) Format(f Formatter) (string, error) {
	var all []property
	for _, v := range p.p {
		all = append(all, *v)
	}
	return f.Format(all)
}

func (p *properties) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("# ")
	buffer.WriteString(comment)
	buffer.WriteString("\n")
	for _, v := range p.p {
		buffer.WriteString("\n")
		buffer.WriteString("# ")
		buffer.WriteString(v.description)
		buffer.WriteString("\n")
		buffer.WriteString(v.Key)
		buffer.WriteString(" = ")
		buffer.WriteString(v.Value)
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func (p *properties) Write(w io.Writer) (int, error) {
	return w.Write([]byte(p.String()))
}

func Properties() *properties {
	return &properties{p: map[string]*property{
		gaugeRepositoryURL:      newProperty(gaugeRepositoryURL, "https://downloads.getgauge.io/plugin", "Url to get plugin versions"),
		gaugeUpdateURL:          newProperty(gaugeUpdateURL, "https://downloads.getgauge.io/gauge", "Url for latest gauge version"),
		gaugeTemplatesURL:       newProperty(gaugeTemplatesURL, "https://downloads.getgauge.io/templates", "Url to get templates list"),
		runnerConnectionTimeout: newProperty(runnerConnectionTimeout, "30000", "Timeout in milliseconds for making a connection to the language runner."),
		pluginConnectionTimeout: newProperty(pluginConnectionTimeout, "10000", "Timeout in milliseconds for making a connection to plugins."),
		pluginKillTimeOut:       newProperty(pluginKillTimeOut, "4000", "Timeout in milliseconds for a plugin to stop after a kill message has been sent."),
		runnerRequestTimeout:    newProperty(runnerRequestTimeout, "30000", "Timeout in milliseconds for requests from the language runner."),
		checkUpdates:            newProperty(checkUpdates, "true", "Allow Gauge and its plugin updates to be notified."),
		telemetryEnabled:        newProperty(telemetryEnabled, "false", "Allow Gauge to collect anonymous usage statistics"),
		telemetryLoggingEnabled: newProperty(telemetryLoggingEnabled, "false", "Log request sent to Gauge telemetry engine"),
	}}
}

func MergedProperties() *properties {
	p := Properties()
	config, err := common.GetGaugeConfiguration()
	if err != nil {
		return p
	}
	for k, v := range config {
		p.set(k, v)
	}
	return p
}

func Update(name, value string) error {
	p := MergedProperties()
	err := p.set(name, value)
	if err != nil {
		return err
	}
	return writeConfig(p)
}

func UpdateTelemetry(value string) error {
	return Update(telemetryEnabled, value)
}

func UpdateTelemetryLoggging(value string) error {
	return Update(telemetryLoggingEnabled, value)
}

func Merge() error {
	return writeConfig(MergedProperties())
}

func GetProperty(name string) (*property, error) {
	return MergedProperties().get(name)
}

func newProperty(key, defaultValue, description string) *property {
	return &property{
		Key:          key,
		defaultValue: defaultValue,
		description:  description,
		Value:        defaultValue,
	}
}

func writeConfig(p *properties) error {
	dir, err := common.GetConfigurationDir()
	f, err := os.OpenFile(filepath.Join(dir, common.GaugePropertiesFile), os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = p.Write(f)
	return err
}