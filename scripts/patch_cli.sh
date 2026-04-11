old_code='        // Combine base discovery commands with artifact isolation flags
        args := \[\]string{scanType, "-t", target}
        if scanType == "scan" {
                args = append(args, "--pt-json", "--out-dir", tmpDir)
        } else if scanType == "specter" {
                args = append(args, "--out-dir", tmpDir)
        } else if scanType == "audit" {
                logFile := filepath.Join(tmpDir, "audit.jsonl")
                args = append(args, "--log-file", logFile)
        } else if scanType == "weirdpackets" {
                // weirdpackets does not output JSON or accept --out-dir
        } else {
                args = append(args, "--json", "--out-dir", tmpDir)
        }'

awk '
BEGIN {
  find=1
}
/Combine base discovery commands/ {
    print "        // Combine base discovery commands with artifact isolation flags\n        var args []string\n        if scanType == \"scan\" {\n                // Map \"scan\" action to the \"recon\" engine for single IP targeting\n                args = []string{\"recon\", \"-i\", target, \"--pt-json\", \"--out-dir\", tmpDir}\n        } else if scanType == \"specter\" {\n                args = []string{\"specter\", \"-t\", target, \"--out-dir\", tmpDir}\n        } else if scanType == \"audit\" {\n                logFile := filepath.Join(tmpDir, \"audit.jsonl\")\n                args = []string{\"audit\", \"-t\", target, \"--log-file\", logFile}\n        } else if scanType == \"weirdpackets\" {\n                // weirdpackets does not output JSON or accept --out-dir\n                args = []string{\"weirdpackets\", \"-t\", target}\n        } else {\n                args = []string{scanType, \"-t\", target, \"--json\", \"--out-dir\", tmpDir}\n        }"
    skip=13
    next
}
{
   if (skip > 0) {
      skip--
   } else {
      print $0
   }
}
' executor/cli.go > executor/cli_new.go
mv executor/cli_new.go executor/cli.go
gofmt -w executor/cli.go
