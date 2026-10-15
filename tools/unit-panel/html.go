package main

import (
	"fmt"
	"html"
	"strings"
)

func esc(s string) string { return html.EscapeString(s) }
func productHTML(draws []Draw) string {
	var b strings.Builder
	b.WriteString("<div class=\"panel\" role=\"group\" aria-label=\"Unit information · Standard rules\">\n")
	for _, r := range draws {
		switch r.Kind {
		case "image":
			alt := ""
			if strings.HasPrefix(r.Input, "exports/portrait") {
				alt = "Infantry portrait"
			}
			if strings.HasPrefix(r.Input, "exports/unit") {
				alt = "Tactical ownership thumbnail"
			}
			srcset := ""
			if strings.HasSuffix(r.Path, ".2x.png") {
				srcset = fmt.Sprintf(" srcset=\"%s 2x, %s 4x\"", esc(r.Path), esc(strings.TrimSuffix(r.Path, ".2x.png")+".4x.png"))
			} else if strings.HasSuffix(r.Path, "-2x.png") {
				srcset = fmt.Sprintf(" srcset=\"%s 2x, %s 4x\"", esc(r.Path), esc(strings.TrimSuffix(r.Path, "-2x.png")+"-4x.png"))
			}
			fmt.Fprintf(&b, "<img src=\"%s\"%s alt=\"%s\" style=\"left:%gpx;top:%gpx;width:%gpx;height:%gpx\">\n", esc(r.Path), srcset, alt, r.X, r.Y, r.Width, r.Height)
		case "text":
			full := r.Full
			if full == "" {
				full = r.Text
			}
			weight := "400"
			if r.Font == "bold" {
				weight = "700"
			}
			fmt.Fprintf(&b, "<span title=\"%s\" aria-label=\"%s\" style=\"left:%gpx;top:%gpx;font-size:%gpx;line-height:%gpx;font-weight:%s;color:%s\">%s</span>\n", esc(full), esc(full), r.X, r.HTMLTop, r.Size, r.Size, weight, esc(r.Color), esc(r.Text))
		}
	}
	b.WriteString("</div>\n")
	return b.String()
}
func galleryHTML(fs []Fixture, renders []FixtureRender) string {
	var b strings.Builder
	b.WriteString(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Unit panel v1 · infantry composition pilot</title>
<style>
@font-face{font-family:HLRGo;src:url('sources/regular.ttf') format('truetype');font-weight:400;font-display:block}
@font-face{font-family:HLRGo;src:url('sources/bold.ttf') format('truetype');font-weight:700;font-display:block}
*{box-sizing:border-box}body{margin:0;padding:24px;background:#09151d;color:#eee5d2;font-family:HLRGo,sans-serif}h1{font-size:24px;margin:0 0 12px}p{max-width:950px;line-height:1.5}.gallery{display:grid;grid-template-columns:repeat(auto-fit,minmax(400px,1fr));gap:24px;max-width:1248px}.card{min-width:0}.scroller{overflow:auto;max-width:100%}.panel{position:relative;width:400px;height:640px;flex:none}.panel img,.panel span{position:absolute}.panel span{white-space:pre}h2{font-size:16px;margin:16px 0 8px}.fixture-note{min-height:44px;font-size:13px;color:#bcc9cc;margin:0 0 8px}.help{font-size:14px;color:#bcc9cc}@media(max-width:450px){body{padding:12px}.gallery{grid-template-columns:minmax(0,1fr)}}
</style></head><body>
<h1>Infantry inspector · Standard</h1>
<p>New unit panel v1 composition pilot. The panel shows current game state from the embedded unit and weapon tables. These static fixtures demonstrate presentation; they provide no attack command.</p>
<p class="help">Weapon class counts and row classes show raw Air / Ground / Water definition flags. Under Standard rules, a Ground weapon can also attack vessels. Capture permits entry into enemy or neutral buildings. Size and weight are abstract values; fuel is the stored number, speed is base speed, and rank is a number from 0 to 10.</p>
<div class="gallery">
`)
	for i, f := range fs {
		fmt.Fprintf(&b, "<section class=\"card\"><h2>%s</h2><p class=\"fixture-note\">%s</p><div class=\"scroller\">%s</div></section>\n", esc(f.ID), esc(f.Label), productHTML(renders[i*2].Draws))
	}
	b.WriteString("</div></body></html>\n")
	return b.String()
}
