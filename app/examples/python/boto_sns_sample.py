import boto
import boto.sqs
import boto.sns

"""
Integration test for boto using GoAws SNS interface
    - Create a virtual environment (pyvenv venv)
    - Activate the venv (source venv/bin/activate)
    - Install boto (pip install boto)
    - run this script (python boto_sns_integration_tests.py)
"""

"""
boto doesn't (yet) expose SetSubscriptionAttributes, so here's a
monkeypatch specifically for turning on the RawMessageDelivery attribute.
"""

def SetRawSubscriptionAttribute(snsConnection, subscriptionArn):
   """
   Works around boto's lack of a SetSubscriptionAttributes call.
   """
   params = {
    'AttributeName': 'RawMessageDelivery',
    'AttributeValue': 'true',
    'SubscriptionArn': subscriptionArn
   }
   return snsConnection._make_request('SetSubscriptionAttributes', params)

boto.sns.SNSConnection.set_raw_subscription_attribute = SetRawSubscriptionAttribute


# Connect GOAws in Python
endpoint='localhost'
region = boto.sqs.regioninfo.RegionInfo(name='local', endpoint=endpoint)
conn = boto.connect_sns(aws_access_key_id='x', aws_secret_access_key='x', is_secure=False, port='4100', region=region)


# Get all Topics in Python
print(conn.get_all_topics())
print()
print()


# Get all Subscriptions in Python
print(conn.get_all_subscriptions())
print()
print()


# Create a topic in Python
topicname = "trialBotoTopic"
topicarn = conn.create_topic(topicname)
print(topicname, "has been successfully created with a topic ARN of", topicarn)
print()
print()


# Print the topic Arn in python
print(topicarn['Result']['TopicArn'])
print()
print()


## Subscribe a Queue to a Topic in Python
subscription1 = conn.subscribe(topicarn['Result']['TopicArn'], "sqs", "http://localhost:4100/queue/local-queue2")
print(subscription1['Result']['SubscriptionArn'])
print()
print()

## Set topic attribute Raw in Python
attr_results = conn.set_raw_subscription_attribute(subscription1['Result']['SubscriptionArn'])
print(attr_results)
print()
print()


## Publish to a topic in Python
message = "Hello Boto"
message_subject = "trialBotoTopic"
publication = conn.publish(topicarn['Result']['TopicArn'], message, subject=message_subject)
print(publication)


## Unsubscribe in Python
subscription1 = conn.unsubscribe(subscription1['Result']['SubscriptionArn'])
print(subscription1)
print()
print()


## Delete Topic in Python
deletion1 = conn.delete_topic(topicarn['Result']['TopicArn'])
print(deletion1)
print()
print()
