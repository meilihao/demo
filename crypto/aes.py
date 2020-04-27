# -*- coding: utf-8 -*-
# import os
import base64
import os
from Crypto.Cipher import AES
import struct
import hashlib

key = "abcwJZhsVZNV0nxogyEuLOCFuKeyZm3A".encode("utf-8")
iv = hashlib.sha256(key).digest()[:16]

def aesEncrypt(key, iv, data):
    def pkcs7padding(data):
        bs = AES.block_size
        padding = bs - len(data) % bs
        padding_text = chr(padding) * padding
        return data + padding_text.encode('utf-8')

    data = pkcs7padding(data)
    cipher = AES.new(key, AES.MODE_CBC, iv)
    encrypted = cipher.encrypt(data)
    return iv+encrypted


def aesDecrypt(key, data):
    def pkcs7unpadding(data):
        length = len(data)
        unpadding = data[length - 1]
        return data[0:length - unpadding]

    cipher = AES.new(key, AES.MODE_CBC, data[:16])
    decrypted = cipher.decrypt(data[16:])
    decrypted = pkcs7unpadding(decrypted)

    return str(decrypted, encoding='utf-8')

if __name__ == "__main__":
    myword = "i am learning python, 但我更喜欢golang"
    encryptedData = aesEncrypt(key, iv, myword. encode('utf-8'))
    plain = aesDecrypt(key, encryptedData)
    if plain == myword:
        print("aes ok")
    else:
        print(myword, plain)

# FAQ:
# 1. No module named 'Crypto' : sudo pip3 install pycryptodome
# 1. Unicode-objects must be encoded before hashing : key + `.encode("utf-8")`
# 1. Object type <class 'str'> cannot be passed to C code : AES.new().encrypt(data.encode('utf-8'))
# 1. ord() expected string of length 1, but int found : ord()需要一个长度为1的字符串, 但`data[length - 1]`已是整数
