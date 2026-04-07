with open("main.go", "r") as f:
    text = f.read()

text = text.replace('} else {\\n\\t\\t\\t\\t\\t\\tlog.Printf("ERROR: Failed parsing SDP', '}\\n\\t\\t\\t\\t\\t} else {\\n\\t\\t\\t\\t\\t\\tlog.Printf("ERROR: Failed parsing SDP')

with open("main.go", "w") as f:
    f.write(text)
