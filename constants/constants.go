package constants

import "strings"

const DefaultCompressionThreshold = int64(1024)
const DefaultCacheSize = 1024 * 1024

var CspTemplate string = strings.Join([]string{
	"default-src 'self' ${_CSP_STYLE_SRC};",
	"connect-src 'self' ${_CSP_CONNECT_SRC};",
	"font-src 'self' ${_CSP_FONT_SRC};",
	"img-src 'self' ${_CSP_IMG_SRC};",
	"script-src 'self' ${NGSS_CSP_NONCE} ${NGSS_CSP_SCRIPT_HASH} ${_CSP_SCRIPT_SRC};",
	"style-src 'self' ${NGSS_CSP_NONCE} ${NGSS_CSP_STYLE_HASH} ${_CSP_STYLE_SRC};",
}, " ")
