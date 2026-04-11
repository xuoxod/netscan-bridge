#!/bin/bash
cat $HOME/private/projects/desktop/golang/netscan_bridge/main.go | awk '
/log.Println\("✅ Execution complete. Streaming output back..."\)/ {
    print "                                        log.Printf(\"✅ Execution complete. Streaming output back... Payload size: %d\", len(out))"
    print "                                        resp := map[string]interface{}{"
    print "                                                \"event\": \"scan_complete\","
    print "                                                \"type\":  scanType,"
    print "                                                \"data\":  out,"
    print "                                        }"
    print "                                        b, _ := json.Marshal(resp)"
    print "                                        bSize := len(b)"
    print "                                        if bSize > 16384 {"
    print "                                                chunkSize := 16384"
    print "                                                totalChunks := (bSize + chunkSize - 1) / chunkSize"
    print "                                                id := fmt.Sprintf(\"%d\", time.Now().UnixNano())"
    print "                                                for i := 0; i < totalChunks; i++ {"
    print "                                                        end := (i + 1) * chunkSize"
    print "                                                        if end > bSize {"
    print "                                                                end = bSize"
    print "                                                        }"
    print "                                                        chunkMsg := map[string]interface{}{"
    print "                                                                \"event\": \"chunk\","
    print "                                                                \"id\":    id,"
    print "                                                                \"index\": i,"
    print "                                                                \"total\": totalChunks,"
    print "                                                                \"data\":  base64.StdEncoding.EncodeToString(b[i*chunkSize : end]),"
    print "                                                        }"
    print "                                                        cb, _ := json.Marshal(chunkMsg)"
    print "                                                        if err := d.SendText(string(cb)); err != nil {"
    print "                                                                log.Printf(\"ERROR sending chunk %d: %v\", i, err)"
    print "                                                        }"
    print "                                                        time.Sleep(5 * time.Millisecond)"
    print "                                                }"
    print "                                        } else {"
    print "                                                if err := d.SendText(string(b)); err != nil {"
    print "                                                        log.Printf(\"ERROR sending scan_complete: %v\", err)"
    print "                                                }"
    print "                                        }"
    skip = 1
    next
}
skip && /d.SendText\(string\(b\)\)/ {
    skip = 0
    next
}
skip && /{/ { skip++; next }
skip && /}/ { skip--; if(skip == 0) { print "                                }()" }; next }
{
    print $0
}
' > $HOME/private/projects/desktop/golang/netscan_bridge/main.go.patched

mv $HOME/private/projects/desktop/golang/netscan_bridge/main.go.patched $HOME/private/projects/desktop/golang/netscan_bridge/main.go
