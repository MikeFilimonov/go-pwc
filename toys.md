port by default and missing port never take place in normalized stored url, while an arbitrary port does

every piece of the key is now decided: Scheme (lowercased), Host (lowercased), Port (dropped if default-or-missing, kept as literal otherwise), Path (trailing slash stripped except root), RawQuery (kept byte-for-byte), Fragment (dropped entirely)


1) 
-  accept the page
- fetch the page
- parse HTML
- collect urls queued bfs
- normalize urls
- collect  stats: page count, found urls, total size
- deduping

2)

 -- worker pool concurrency
 --  rate limiting
 -- graceful shutdown

3) contect cancellation:

--- crawling termination from the outside world
--- limiting the total execution time
--- resource flushing on cancellation