import yaml
import os
import argparse
import subprocess # Added for gofmt

def to_pascal_case(components):
    return "".join(x.replace("_", " ").title().replace(" ", "") for x in components)

def to_upper_snake_case(components):
    return "_".join(x.upper() for x in components)

def walk_yaml(data, path_list):
    results = []
    if isinstance(data, dict):
        for key, value in data.items():
            results.extend(walk_yaml(value, path_list + [key]))
    elif isinstance(data, list):
        for item in data:
            results.append(path_list + [item])
    else:
        results.append(path_list + [str(data)])
    return results

def write_file(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as f:
        f.write(content)

def main():
    parser = argparse.ArgumentParser(description="Generate Kafka topic constants from YAML.")
    parser.add_argument("--input", required=True, help="Path to topics.yaml")
    parser.add_argument("--go", required=True, help="Full path for topics.go")
    parser.add_argument("--py", required=True, help="Full path for topics.py")
    parser.add_argument("--js", required=True, help="Full path for topics.js")
    
    args = parser.parse_args()

    if not os.path.exists(args.input):
        print(f"Error: Input file {args.input} not found.")
        return

    with open(args.input, 'r') as f:
        content = yaml.safe_load(f)
    
    raw_topics = walk_yaml(content.get('topics', {}), [])
    
    # Use a single tab (\t) - gofmt will handle the alignment perfectly
    go_lines = ["package event", "", "type Topic string", "", "const ("]
    py_lines = []
    js_lines = []

    for parts in raw_topics:
        topic_value = ".".join(parts)
        
        go_const = to_pascal_case(parts)
        go_lines.append(f'\t{go_const} Topic = "{topic_value}"')
        
        upper_const = to_upper_snake_case(parts)
        py_lines.append(f'{upper_const} = "{topic_value}"')
        js_lines.append(f'export const {upper_const} = "{topic_value}";')

    go_lines.append(")")

    # Write files
    write_file(args.go, "\n".join(go_lines))
    write_file(args.py, "\n".join(py_lines))
    write_file(args.js, "\n".join(js_lines))

    # Trigger gofmt to fix the alignment to Go standards automatically
    try:
        subprocess.run(["gofmt", "-w", args.go], check=True)
        print(f"✨ Formatted {args.go} with gofmt")
    except FileNotFoundError:
        print("⚠️  Warning: gofmt not found in PATH. Go file might not be aligned.")

    print(f"✅ Success! Generated constants in:\n - {args.go}\n - {args.py}\n - {args.js}")

if __name__ == "__main__":
    main()