import com.amazonaws.AmazonClientException;
import com.amazonaws.AmazonServiceException;
import com.amazonaws.auth.BasicAWSCredentials;
import com.amazonaws.services.sns.model.*;
import com.amazonaws.services.sns.AmazonSNS;
import com.amazonaws.services.sns.AmazonSNSClient;

import java.util.List;
import java.util.Map.Entry;

/***
 * Make sure you have the aws-java-sdk-1.8.11.jar + dependancies in your classpath
***/

public class SnsSample {

    public static void main(String[] args) throws Exception {

        AmazonSNS sns = new AmazonSNSClient(new BasicAWSCredentials("x", "x"));
        sns.setEndpoint("http://localhost:4100");

        System.out.println("===========================================");
        System.out.println("Getting Started with Amazon SQS");
        System.out.println("===========================================\n");

        try {
            // Create a queue
            System.out.println("Creating a new SNS topic called MyTopic.\n");
            CreateTopicRequest createTopicRequest = new CreateTopicRequest("MyTopic");
            String topicArn = sns.createTopic(createTopicRequest).getTopicArn();

            // List queues
            System.out.println("Listing all topics in your account.\n");
            for (Topic topic : sns.listTopics().withTopics().getTopics()) {
                System.out.println("  TopicArn: " + topic.getTopicArn());
            }
            System.out.println();

            SubscribeResult sr = sns.subscribe(new SubscribeRequest(topicArn, "sqs", "http://localhost:4100/queue/local-queue1"));
            System.out.println("SubscriptionArn: " + sr.getSubscriptionArn());
            System.out.println();

            PublishRequest publishRequest = new PublishRequest(topicArn, "Sent to MyTopic!!!");
            PublishResult pr = sns.publish(publishRequest);
            System.out.println("Message sent: " + pr.getMessageId());
            System.out.println();

            DeleteTopicRequest str = new DeleteTopicRequest();
            str.setTopicArn(topicArn);
            sns.deleteTopic(str);
            System.out.println("Topic Delected: " + topicArn);
            System.out.println();
       } catch (AmazonServiceException ase) {
            System.out.println("Caught an AmazonServiceException, which means your request made it " +
                    "to Amazon SQS, but was rejected with an error response for some reason.");
            System.out.println("Error Message:    " + ase.getMessage());
            System.out.println("HTTP Status Code: " + ase.getStatusCode());
            System.out.println("AWS Error Code:   " + ase.getErrorCode());
            System.out.println("Error Type:       " + ase.getErrorType());
            System.out.println("Request ID:       " + ase.getRequestId());
        } catch (AmazonClientException ace) {
            System.out.println("Caught an AmazonClientException, which means the client encountered " +
                    "a serious internal problem while trying to communicate with SQS, such as not " +
                    "being able to access the network.");
            System.out.println("Error Message: " + ace.getMessage());
        }
    }
}
