


```dataview
TABLE WITHOUT ID authors as "Authors", round(sum(map(rows, (r) => default(r.file.size, 0)))/6)  as "Words Approx.", sum(map(rows, (r) => default(r.file.size, 0))) as "Size (bytes)"
FROM "Entries"
WHERE process_step != null
FLATTEN authors
GROUP By authors
SORT sum(map(rows, (r) => default(r.file.size, 0)))
```


prooferead
```dataview
TABLE WITHOUT ID proofread_by as "proofreader", round(sum(map(rows, (r) => default(r.file.size, 0)))/6)  as "Words Approx.", sum(map(rows, (r) => default(r.file.size, 0))) as "Size (bytes)"
FROM "Entries"
WHERE process_step != null
FLATTEN proofread_by
GROUP By proofread_by
SORT sum(map(rows, (r) => default(r.file.size, 0)))
```


**REALLY APPROXIMATE NUMBERS**
```js
let pages = dv.pages('"Entries/Hardware Entries"').filter(d => d.notebook === "hardware");
let entriesWithManyAuthors = pages.filter(f => f.authors != null && f.authors.length > 1);

let bytes = pages.map(f => f.file.size).sum();

let words = Math.floor(bytes/6);
dv.span(`Hardware: ${bytes} bytes. ~${words} words. ${entriesWithManyAuthors.length} entries with overlap authors`);

dv.span(entriesWithManyAuthors.file.link);

dv.span("Authors")
dv.span([...new Set(pages.map(d => d.authors == null ? "" : d.authors[0]))].sort())
```

```js
let pages = dv.pages('"Entries/Software Entries"').filter(d => d.notebook === "software");
let entriesWithManyAuthors = pages.filter(f => f.authors.length > 1);

let bytes = pages.map(f => f.file.size).sum();

let words = Math.floor(bytes/6);
dv.span(`Software: ${bytes} bytes. ${words} ~words. ${entriesWithManyAuthors.length} entries with overlap authors`);

dv.span(entriesWithManyAuthors.file.link);

dv.span("Authors")
dv.span([...new Set(pages.map(d => d.authors[0]))].sort())
```
 
```js
let pages = dv.pages('"Entries/Strategy Entries"').filter(d => d.notebook === "strategy");
let entriesWithManyAuthors = pages.filter(f => f.authors != null && f.authors.length > 1);

let bytes = pages.map(f => f.file.size).sum();

let words = Math.floor(bytes/6);
dv.span(`Straregy: ${bytes} bytes. ~${words} words. ${entriesWithManyAuthors.length} entries with overlap authors`);

dv.span(entriesWithManyAuthors.file.link);

dv.span("Authors")
dv.span([...new Set(pages.map(d => d.authors == null ? "" : d.authors[0]))].sort())```


## File Sizes

```dataview
TABLE file.size as "File Size (b)", authors as "Authors"
FROM "Entries"
WHERE process_step != null
SORT file.size desc
```
