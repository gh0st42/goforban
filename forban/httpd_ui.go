package forban

import "github.com/dustin/go-humanize"

var htmlheader = `
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en"> <head> <link rel="stylesheet" type="text/css" href="/assets/x.css"
/></head>
`

var htmlfooter = `
<div id="w3c"><p><small>
Forban is free software released under the <a href="http://www.gnu.org/licenses/agpl.html">AGPL</a>. For more information about Forban and source code : <a href="http://www.foo.be/forban/">foo.be/forban/</a>.
</small></p></div></div><!-- end wrapper --></div>
</body></html>`

func htmlnav() string {
	return `<body><div id="nav"><a href="/"><img src="assets/forban-small.png" alt="forban
logo : a small island where a stream of bits is going to and coming from"
/></a><br /><ul><li><span class="home">Description : <i>` + MyName + `</i><br/>Mode : <i>Opportunistic</i></span></li>
</ul></div><div id="wrapper">`
}

func getindexhtml() string {
	html := htmlheader
	html += htmlnav()
	html += `<br/> <br/> <br/> <div class="right inner">`
	html += `<h2>Search the loot...</h2>`
	html += `<form method=get action="q/"><input type="text" name="v" value=""> <input
        type="submit" value="search"></form>`
	html += `</div> <div class="left inner">`
	html += `<h2>Discovered link-local Forban available with their loot in the last 3 minutes</h2> `
	html += `<table>`
	html += `<th><td>Access</td><td>Name</td><td>Last seen</td><td>First seen</td><td>Size</td><td>How many files are missing from yourself?</td><td></td></th>`

	for _, value := range Neighborhood {
		html += `<tr>`
		//fmt.Fprintf(w, "%s \t %s \t %s \t %s \t %s\n",
		//	key, value.node.name, value.node.hmac,
		//	time.Since(value.firstSeen), time.Since(value.node.lastSeen))

		if value.node.ipv4 != "" {
			html += `<td><a href="http://` + value.node.ipv4 + `:12555/">v4</a></td> `
		} else {
			html += `<td></td> `
		}
		if value.node.ipv6 != "" {
			html += `<td><a href="http://` + value.node.ipv4 + `:12555/">v4</a></td> `
		} else {
			html += `<td></td> `
		}
		html += `<td>` + value.node.name + `</td> `
		html += `<td>` + humanize.Time(value.node.lastSeen) + `</td> `
		html += `<td>` + humanize.Time(value.firstSeen) + `</td> `
		//		html += `<td>` + time.Since(value.node.lastSeen).String() + `</td> `
		html += `<td>` + humanize.Bytes(uint64(value.totalStore)) + `</td> `

		html += `</tr>`
	}

	html += `</table></div>`
	html += htmlfooter
	return html
}
