Writing is difficult, we should probably proofread it. Below are lists of entries that have been marked **finished** but haven't been proofread by anyone yet 

If you spot a simple grammar mistake, feel free to fix it. But if there's something big that needs clearing up that you aren't able to fix, ping the author on slack. Slack messages are free 👍

```dataview
TABLE authors, notebook, proofread_by 
FROM "Entries" 
WHERE finished = true SORT length(proofread_by)
```


