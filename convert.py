import json
import re

# Path to the input file
input_file_path = "cedict_ts.u8"  # Replace with the actual path to your file

# Paths to the output JSON files
output_json_path = "cedict2.json"  # Replace with the desired output path
simplified_to_traditional_path = (
    "simplified_to_traditional.json"  # Replace with the desired output path
)
traditional_list_path = "traditional_list.json"  # Replace with the desired output path

# Dictionaries to hold the parsed data
cedict_dict = {}
simplified_to_traditional_dict = {}
traditional_list = []

# Regular expression to parse each line of the file
pattern = re.compile(r"(\S+) (\S+) \[(.+?)\] /(.+)/")

with open(input_file_path, "r", encoding="utf-8") as file:
    for line in file:
        # Skip comments
        if line.startswith("#"):
            continue

        match = pattern.match(line)
        if match:
            traditional, simplified, pinyin, definitions = match.groups()

            # Add to the list of traditional characters
            traditional_list.append(traditional)

            # Create entry in the cedict_dict
            cedict_dict[traditional] = {
                "traditional": traditional,
                "simplified": simplified,
                "pinyin": pinyin.split(),
                "definitions": definitions.split("/"),
            }

            # Create entry in the simplified_to_traditional_dict
            simplified_to_traditional_dict[simplified] = traditional

# Write the dictionaries and list to their respective JSON files
with open(output_json_path, "w", encoding="utf-8") as file:
    json.dump(cedict_dict, file, ensure_ascii=False, indent=4)

with open(simplified_to_traditional_path, "w", encoding="utf-8") as file:
    json.dump(simplified_to_traditional_dict, file, ensure_ascii=False, indent=4)

with open(traditional_list_path, "w", encoding="utf-8") as file:
    json.dump(traditional_list, file, ensure_ascii=False, indent=4)
