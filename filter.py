import os
import json

file = open("./ids.json", "r", encoding="utf-16")

string = file.read().replace('\x00', '')

obj = json.loads(string)

shipGroups = [
  25, 26, 27, 28, 29, 30, 31, 237, 324, 358, 380, 381, 419, 420, 463, 485,
  513, 540, 541, 543, 547, 659, 830, 831, 832, 833, 834, 883, 893, 894, 898,
  900, 902, 906, 941, 963, 1022, 1201, 1202, 1283, 1305, 1527, 1534, 1538,
  1972, 2001, 4594,
]

ships = {}

for key in obj:
  if obj[key]["groupID"] in shipGroups:
    ships[key] = obj[key]

stringified = json.dumps(ships)

ships = open("./ships.json", "w")

ships.write(stringified)

