; --- "even" data directive test ---
start:
    dc.b    $11, $22, $33
    even
    dc.w    $4455
    dc.b    $66
    even
    dc.l    $778899aa

gap:
    ds.b    3
    dc.b    $ff
