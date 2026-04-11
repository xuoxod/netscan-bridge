old_code="            if (scanType === 'scan') {
                const protocol = document.getElementById('studioProtocol') ? document.getElementById('studioProtocol').value : 'tcp';
                const ports = document.getElementById('studioPorts') ? document.getElementById('studioPorts').value.trim() : '1-1000';

                flags.push('--protocol');
                flags.push(protocol);
                if (ports) {
                    flags.push('-p');
                    flags.push(ports);
                }
                resetBtnText = \"Launch Deep Recon\";
            } else if (scanType === 'audit') {
                const duration = document.getElementById('auditDuration') ? document.getElementById('auditDuration').value.trim() : '15';
                const spoof = document.getElementById('auditSpoof') ? document.getElementById('auditSpoof').checked : false;
                const stealth = document.getElementById('auditStealth') ? document.getElementById('auditStealth').checked : false;

                if (duration) {
                    flags.push('--duration');
                    flags.push(duration);
                }
                if (spoof) flags.push('--spoof');
                if (stealth) flags.push('--stealth');
                resetBtnText = \"Execute Privacy Audit\";
            } else if (scanType === 'weirdpackets') {
                const count = document.getElementById('wpCount') ? document.getElementById('wpCount').value.trim() : '100';
                if (count) {
                    flags.push('--count');
                    flags.push(count);
                }
                resetBtnText = \"Stress Test Firewall\";
            }"

awk '
BEGIN { skip = 0 }
/if \(scanType === .\scan.\)/ {
    print "            if (scanType === \"scan\") {\n                const protocol = document.getElementById(\"studioProtocol\") ? document.getElementById(\"studioProtocol\").value : \"tcp\";\n                const ports = document.getElementById(\"studioPorts\") ? document.getElementById(\"studioPorts\").value.trim() : \"1-1000\";\n\n                if (protocol === \"udp\" || protocol === \"both\") {\n                    flags.push(\"--udp-open-filtered\"); // recon engine maps protocol this way\n                }\n                if (ports) {\n                    flags.push(\"-p\");\n                    flags.push(ports);\n                }\n                resetBtnText = \"Launch Deep Recon\";\n            } else if (scanType === \"audit\") {\n                const duration = document.getElementById(\"auditDuration\") ? document.getElementById(\"auditDuration\").value.trim() : \"15\";\n                const spoof = document.getElementById(\"auditSpoof\") ? document.getElementById(\"auditSpoof\").checked : false;\n                const stealth = document.getElementById(\"auditStealth\") ? document.getElementById(\"auditStealth\").checked : false;\n\n                if (duration) {\n                    flags.push(\"-d\"); // Using short flag for duration\n                    flags.push(duration);\n                }\n                if (spoof) flags.push(\"--spoof\");\n                if (stealth) flags.push(\"--stealth\");\n                resetBtnText = \"Execute Privacy Audit\";\n            } else if (scanType === \"weirdpackets\") {\n                // weirdpackets currently does not support --count in the Rust CLI, so we will pass empty extra flags\n                resetBtnText = \"Stress Test Firewall\";"
    skip = 29
    next
}
{
    if (skip > 0) {
        skip--
    } else {
        print $0
    }
}
' ../rmediatech/static/js/live-recon.js > tmp_live.js
mv tmp_live.js ../rmediatech/static/js/live-recon.js
