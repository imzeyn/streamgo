package streamgo

import "strings"

func ClearURL(s string) string {
    var builder strings.Builder
    builder.WriteString("/")

    for _, v := range strings.Split(s, "/") {
        if strings.TrimSpace(v) != "" {
            builder.WriteString(v + "/")
        }
    }
	
    return builder.String()
}
