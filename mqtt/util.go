package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/xBlaz3kx/DevX/configuration"
	"github.com/xBlaz3kx/DevX/observability"
	"go.uber.org/zap"
)

var (
	ErrNotValidSubscriptionTopic = fmt.Errorf("not a valid subscription topic")
	ErrNotSameTopic              = fmt.Errorf("not the same topic")
	ErrNotSubscribedTopic        = fmt.Errorf("not the subscribed topic")
	ErrInvalidArgs               = fmt.Errorf("invalid number of arguments")
	ErrInvalidIds                = fmt.Errorf("ids cannot be an empty string")
)

// GetIdsFromTopic parses the topic received from the MQTT client and returns the ids based on the original subscription topic.
// For example:
// actual topic = some/exampleId1/subscription/exampleId2/topic
// subscription topic = some/+/subscription/+/topic
// should return ["exampleId1", "exampleId2"]
// If the topic are not the same length or don't contain the same words, it will return an error
func GetIdsFromTopic(obs observability.Observability, actualTopic string, subTopic Topic) ([]string, error) {
	obs.Log().With(
		zap.String("actualTopic", actualTopic),
		zap.String("originalTopic", string(subTopic)),
	).Debug("Getting Ids from topic")

	var (
		ids               []string
		actualTopicValues = strings.Split(actualTopic, "/")
		subTopicValues    = strings.Split(subTopic.String(), "/")
	)

	// Check if it is the same length, which would indicate the same topic
	if len(actualTopicValues) != len(subTopicValues) {
		return nil, ErrNotSameTopic
	}

	// Check if the subscription topic has at least one + or #
	if !strings.ContainsAny(subTopic.String(), "+#") {
		return nil, ErrNotValidSubscriptionTopic
	}

	for i, value := range subTopicValues {
		if value != actualTopicValues[i] && value != "+" {
			return nil, ErrNotSubscribedTopic
		} else if value == "+" {
			ids = append(ids, actualTopicValues[i])
		}
	}

	return ids, nil
}

// CreateTopicWithIds replaces all the + sign in a topic used for subscription with ids. Works only if the number of pluses is matches the number of ids.
func CreateTopicWithIds(obs observability.Observability, topicTemplate Topic, ids ...string) (string, error) {
	obs.Log().With(
		zap.String("topic", string(topicTemplate)),
		zap.Strings("ids", ids),
	).Debug("Creating publish topic")

	finalString := topicTemplate.String()

	// Check if the number of pluses is the same
	if strings.Count(topicTemplate.String(), "+") != len(ids) {
		return "", ErrInvalidArgs
	}

	// Any empty string is invalid
	if stringUtils.IsAnyEmpty(ids...) {
		return "", ErrInvalidIds
	}

	for _, id := range ids {
		finalString = strings.Replace(finalString, "+", id, 1)
	}

	return finalString, nil
}

// createTlsConfiguration Create a TLS cert using the private key and certificate
func createTlsConfiguration(obs observability.Observability, tlsSettings configuration.TLS) *tls.Config {
	clientCert, err := tls.LoadX509KeyPair(tlsSettings.CertificatePath, tlsSettings.PrivateKeyPath)
	if err != nil {
		obs.Log().Sugar().Fatalf("invalid key pair: %v", err)
	}

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	certs, err := os.ReadFile(tlsSettings.RootCertificatePath)
	if err != nil {
		obs.Log().Sugar().Fatalf("Failed to append to RootCAs: %v", err)
	}

	if !rootCAs.AppendCertsFromPEM(certs) {
		obs.Log().Sugar().Error("No certs appended, using system certs only")
	}

	return &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            rootCAs,
		Certificates: []tls.Certificate{
			clientCert,
		},
	}
}
