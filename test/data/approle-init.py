#!/usr/bin/env python3

import json
import urllib.request
import os

url = os.environ['VAULT_ADDR']
token = os.environ['VAULT_TOKEN']
role = os.environ['VAULT_ROLE']
auth_path = os.environ['VAULT_AUTH_PATH']
id_store = os.environ['ID_STORE']


role_id_url = "{0}/v1/auth/{1}/role/{2}/role-id".format(url,auth_path,role)
role_id_path = "{0}/role-id".format(id_store)
secret_id_url = "{0}/v1/auth/{1}/role/{2}/secret-id".format(url,auth_path,role)
secret_id_path = "{0}/initial-secret-id".format(id_store)
headers = {
    'X-Vault-Token': token
}

print("Requesting new role-id from",role_id_url)
req = urllib.request.Request(role_id_url,headers=headers)
response = urllib.request.urlopen(req)
data = json.loads(response.read().decode('utf8'))

if 'errors' in data:
    raise ValueError(','.join(data['errors']))

role_id = data['data']['role_id']

print("Requesting new secret-id from",secret_id_url)
req = urllib.request.Request(secret_id_url,headers=headers, method='POST')
response = urllib.request.urlopen(req)
data = json.loads(response.read().decode('utf8'))

if 'errors' in data:
    raise ValueError(','.join(data['errors']))

secret_id = data['data']['secret_id']

print("Writing role-id",role_id,"to",role_id_path)
f = open(role_id_path,"w")
f.write(role_id)
f.close()

print("Writing secret-id",secret_id,"to",secret_id_path)
f = open(secret_id_path,"w")
f.write(secret_id)
f.close()