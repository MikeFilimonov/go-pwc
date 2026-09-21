package pwc

port by default and missing port never take place in normalized stored url, while an arbitrary port does

every piece of the key is now decided: Scheme (lowercased), Host (lowercased), Port (dropped if default-or-missing, kept as literal otherwise), Path (trailing slash stripped except root), RawQuery (kept byte-for-byte), Fragment (dropped entirely)