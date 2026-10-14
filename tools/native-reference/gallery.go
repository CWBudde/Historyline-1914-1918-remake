package main

import (
	"bytes"
	"html/template"
	"sort"
)

type galleryItem struct {
	ID, Library, Name, Coverage, Palette, PalettePath, Platform, Path, Aspect, MetadataPath string
	Width, Height                                                                           int
}

func renderGallery(items []galleryItem, hasScreen bool) ([]byte, error) {
	seen := make(map[string]bool)
	var libraries []string
	for _, item := range items {
		if !seen[item.Library] {
			seen[item.Library] = true
			libraries = append(libraries, item.Library)
		}
	}
	sort.Strings(libraries)
	data := struct {
		Items     []galleryItem
		Libraries []string
		HasScreen bool
	}{items, libraries, hasScreen}
	t, err := template.New("gallery").Parse(galleryHTML)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

const galleryHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Native Amiga references</title>
<style>
:root{color-scheme:dark;--surface:#20262e;--line:#46515d;--text:#edf1f5;--muted:#b8c3cf;--accent:#b9d7f5}
*{box-sizing:border-box}body{margin:0;padding:24px;background:#14191f;color:var(--text);font:16px/1.5 system-ui,sans-serif}h1{margin:0 0 8px;font-size:26px}p{max-width:80ch;color:var(--muted)}a{color:var(--accent)}header{margin-bottom:20px}.controls{display:flex;flex-wrap:wrap;gap:16px;padding:16px;background:var(--surface);border:1px solid var(--line);border-radius:6px}label{display:flex;flex-direction:column;gap:4px}input,select{font:inherit;padding:6px 10px;color:var(--text);background:#14191f;border:1px solid var(--line);border-radius:4px}input{width:min(320px,75vw)}#count{color:var(--muted);margin:12px 0}.gallery{display:grid;grid-template-columns:repeat(auto-fill,minmax(min(280px,100%),1fr));gap:16px}.card{min-width:0;padding:16px;border:1px solid var(--line);border-radius:6px;background:var(--surface)}.card:target{outline:3px solid var(--accent)}.card[hidden]{display:none}.card h2{margin:0 0 10px;font-size:18px;overflow-wrap:anywhere}.preview{overflow:auto;max-height:70vh;padding:16px;margin-bottom:12px;background-color:#c4c4c4;background-image:linear-gradient(45deg,#8e8e8e 25%,transparent 25%),linear-gradient(-45deg,#8e8e8e 25%,transparent 25%),linear-gradient(45deg,transparent 75%,#8e8e8e 75%),linear-gradient(-45deg,transparent 75%,#8e8e8e 75%);background-size:16px 16px;background-position:0 0,0 8px,8px -8px,-8px 0}.preview img{display:block;image-rendering:pixelated;max-width:none}.dark .preview{background:#111}.light .preview{background:#fff}.card dl{display:grid;grid-template-columns:auto 1fr;gap:2px 12px;font-size:14px;margin:0 0 10px}.card dt{color:var(--muted)}.card dd{margin:0;overflow-wrap:anywhere}.files{display:flex;flex-wrap:wrap;gap:12px}.screen{grid-column:1/-1}noscript{color:#ffe0a0}
</style>
</head>
<body>
<header>
<h1>Native Amiga references</h1>
<p>Intrinsic-size original pixels for local inspection. Integer magnification adds no authored detail. BMHD aspect fields are recorded without applying a display correction. Optional map images are static ground reconstructions, not game captures. Physical display aspect, remaining semantic mappings and visual approval remain open.</p>
<nav class="files" aria-label="Export files"><a href="metadata.csv">Library metadata CSV</a>{{if .HasScreen}}<a href="screen-metadata.csv">Screen metadata CSV</a>{{end}}<a href="source-sha256.txt">Source SHA-256 manifest</a><a href="HEXMSK00.png">Native hex coverage mask</a></nav>
</header>
<div class="controls">
<label for="search">Search name or library<input id="search" type="search" placeholder="e.g. SFORE150 or PARTS"></label>
<label for="library">Library<select id="library"><option value="">All libraries</option>{{range .Libraries}}<option value="{{.}}">{{.}}</option>{{end}}</select></label>
<label for="scale">Inspection scale<select id="scale"><option value="1">1× intrinsic</option><option value="2" selected>2×</option><option value="4">4×</option><option value="8">8×</option></select></label>
<label for="background">Background<select id="background"><option value="checker">Checkerboard</option><option value="dark">Dark</option><option value="light">Light</option></select></label>
</div>
<p id="count" aria-live="polite">{{len .Items}} references</p>
<noscript>Enable JavaScript for search, filters and inspection scale. All PNGs and metadata links remain available below.</noscript>
<main class="gallery" id="gallery">
{{range .Items}}<article class="card{{if eq .Palette "CMAP"}} screen{{end}}" id="{{.ID}}" data-library="{{.Library}}" data-name="{{.Name}}">
<h2>{{.Library}} / {{.Name}}</h2>
<div class="preview"><img src="{{.Path}}" alt="{{.Library}} / {{.Name}}" width="{{.Width}}" height="{{.Height}}" data-width="{{.Width}}" data-height="{{.Height}}" loading="lazy"></div>
<dl><dt>Intrinsic size</dt><dd>{{.Width}} × {{.Height}}</dd><dt>Coverage</dt><dd>{{.Coverage}}</dd><dt>Palette</dt><dd>{{.Palette}}</dd><dt>Platform</dt><dd>{{.Platform}}</dd>{{if .Aspect}}<dt>BMHD aspect</dt><dd>{{.Aspect}} (metadata only)</dd>{{end}}</dl>
<div class="files"><a href="{{.Path}}">PNG file</a><a href="#{{.ID}}">Permalink</a>{{if .PalettePath}}<a href="{{.PalettePath}}">Palette swatches</a>{{end}}{{if .MetadataPath}}<a href="{{.MetadataPath}}">Ground cell coordinates / raw state</a>{{end}}</div>
</article>{{end}}
</main>
<script>
(function () {
'use strict';
const gallery = document.getElementById('gallery');
const cards = Array.from(gallery.querySelectorAll('.card'));
const search = document.getElementById('search');
const library = document.getElementById('library');
const scale = document.getElementById('scale');
const background = document.getElementById('background');
function filter() {
 const query = search.value.trim().toLowerCase();
 let visible = 0;
 cards.forEach(function (card) {
  const matches = (!library.value || card.dataset.library === library.value) && (card.dataset.library + ' ' + card.dataset.name).toLowerCase().includes(query);
  card.hidden = !matches;
  if (matches) visible++;
 });
 document.getElementById('count').textContent = visible + ' / ' + cards.length + ' references';
}
function magnify() {
 const factor = Number(scale.value);
 if (![1, 2, 4, 8].includes(factor)) return;
 gallery.querySelectorAll('img').forEach(function (img) {
  img.style.width = (Number(img.dataset.width) * factor) + 'px';
  img.style.height = (Number(img.dataset.height) * factor) + 'px';
 });
}
search.addEventListener('input', filter);
library.addEventListener('change', filter);
scale.addEventListener('change', magnify);
background.addEventListener('change', function () {
 gallery.classList.remove('dark', 'light');
 if (background.value === 'dark' || background.value === 'light') gallery.classList.add(background.value);
});
function revealPermalink() {
 let id;
 try { id = decodeURIComponent(location.hash.slice(1)); } catch (_) { return; }
 const target = document.getElementById(id);
 if (target && target.classList.contains('card')) {
  search.value = ''; library.value = ''; filter(); target.scrollIntoView();
 }
}
filter(); magnify(); revealPermalink();
window.addEventListener('hashchange', revealPermalink);
}());
</script>
</body>
</html>
`
