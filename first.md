# Hello

```sh {"name": "print-non-interactive", "interactive": false}
echo "Hello "
read name <name.txt
echo "$name"
```

```sh {"name": "print-interactive"}
echo "Give me your name: "
read name
echo "Hello, $name!"
```

```sh {"name":"python", "interactive": true}
python
```

```sh {"name": "bash"}
bash
```

```sh {"name": "itself"}
runme tui
```
