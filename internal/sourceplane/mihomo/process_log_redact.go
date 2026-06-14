package mihomo

import "regexp"

var processLogRedactors = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization:\s*bearer\s+)[^\s]+`),
	regexp.MustCompile(`(?i)([?&](?:token|secret|password|passwd|api_key|apikey)=)[^&\s]+`),
	regexp.MustCompile(`(?i)\b((?:https?|socks5h?)://)([^/\s:@]+):([^@\s/]+)@`),
}

func redactProcessLogLine(line string) string {
	line = processLogRedactors[0].ReplaceAllString(line, `${1}<redacted>`)
	line = processLogRedactors[1].ReplaceAllString(line, `${1}<redacted>`)
	line = processLogRedactors[2].ReplaceAllString(line, `${1}<redacted>:<redacted>@`)
	return line
}
