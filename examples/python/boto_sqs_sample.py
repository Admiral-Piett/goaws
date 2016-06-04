import boto
import boto.sqs

"""
Integration test for boto using GoAws SNS interface
    - Create a virtual environment (pyvenv venv)
    - Activate the venv (source venv/bin/activate)
    - Install boto (pip install boto)
    - run this script (python boto_sns_integration_tests.py)
"""

# Connect GOAws in Python
endpoint='localhost'
region = boto.sqs.regioninfo.RegionInfo(name='local', endpoint=endpoint)
conn = boto.connect_sqs(aws_access_key_id='x', aws_secret_access_key='x', is_secure=False, port='4100', region=region)


# Get all Queues in Python
print(conn.get_all_queues())
print()
print()

# Create a queue in Python
q = conn.create_queue('myqueue')
print(q)
print()
print()

# Get Queue Attributes in Python
attribs = conn.get_queue_attributes(q)
print(attribs)
print()
print()

# Get A Queue in Python
qi = conn.get_queue('myqueue')
print(qi)
print()
print()

# Lookup a queue in Python (same as get a queue)
qi = conn.lookup('myqueue')
print(qi)
print()
print()

# Send a message to a queue in Python
resp = conn.send_message(qi, "This is a test!!!")
print(resp)
print()
print()

# Receive a message from a queue in Python
resp2 = conn.receive_message(qi)
for result in resp2:
    print(result.get_body())

    # Delete a message from a queue in Python
    resp3 = conn.delete_message(qi, result)
    print("\tDelete:", resp3)

print()
print()

# Delete a queue in Python
dq = conn.delete_queue(q)
print(dq)
print()
print()


