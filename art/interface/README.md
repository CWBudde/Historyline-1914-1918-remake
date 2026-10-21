# Interface source images

Every image family has PNG source masters beside its `prompt.md`. The completed set contains 109
masters: 86 newly generated interface variants, seven matching button and country masters reused
from `art/screens/`, and the existing 15 weapon paintings and explosion atlas. The root `prompt.md`
describes bindings and animation metadata, so it has no image.

The new sources follow the family prompts, slate-and-brass palette and corresponding runtime
layout references. State names become filenames with spaces and slashes replaced by hyphens. The explosion
atlas supplies all nine explosion frames; only the two missing hit variants need separate sources.

These are candidate images at their native imagegen dimensions. They have not been resized or
promoted to runtime assets. Final pixel alignment, visual acceptance and export bindings remain open.
The generated catalogue records master filenames and dimensions, and the art book displays them.
Some sources retain soft shading and low-alpha edge speckles that need cleanup before runtime
promotion. Action buttons, the content panel and filled slot states preserve solid opaque backings;
corner masks remain part of runtime export preparation.

The local [source gallery](../../build/review/interface-sources/index.html) includes every added
variant and its exact generation prompt. Its `sources.json` records native dimensions and alpha
counts; `prompts.json` also identifies the seven reused screen sources.
