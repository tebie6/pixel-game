package conf

import "gopkg.in/ini.v1"

var (
	configFile *ini.File
)

func GetConfigString(section string, name string) string {
	return configFile.Section(section).Key(name).String()
}

func GetConfigInt(section string, name string) (int64, error) {
	return configFile.Section(section).Key(name).Int64()
}

// 解析config。ini
func ParseConfigINI(cpath string) (err error) {

	configFile, err = ini.Load(cpath)
	if err != nil {
		return err
	}

	return nil
}