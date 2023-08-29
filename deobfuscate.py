import argparse
parser = argparse.ArgumentParser()
parser.add_argument('--dumpname', help='Dump name', default="PPLBlade.dmp")
parser.add_argument('--key', help='XOR key', default="PPLBlade")
args = parser.parse_args()

def xor(input_data, key):
    result = bytearray(len(input_data))
    for i in range(len(input_data)):
        result[i] = input_data[i] ^ key[i % len(key)]
    return bytes(result)


key = args.key.strip().encode()
with open(args.dumpname, 'rb') as f:
    encrypted_data = f.read()

decrypted_data = xor(encrypted_data, key)

with open('decrypted.dmp', 'wb') as f:
    f.write(decrypted_data)
print("[+] Deobfuscated dump saved in file decrypted.dmp")
