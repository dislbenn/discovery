package ocm

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/ocm/auth"
	"github.com/open-cluster-management/discovery/pkg/ocm/subscription"
)

var (
	getTokenFunc func(auth.AuthRequest) (string, error)

	getSubscriptionsFunc func() ([]subscription.Subscription, error)
	subscriptionGetter   = subscriptionGetterMock{}
)

// This mocks the authService request and returns a dummy access token
type authServiceMock struct{}

func (m *authServiceMock) GetToken(request auth.AuthRequest) (string, error) {
	return getTokenFunc(request)
}

// The mocks the GetClusters request to return a select few clusters without connection
// to an external datasource
type subscriptionGetterMock struct{}

func (m *subscriptionGetterMock) GetSubscriptions() ([]subscription.Subscription, error) {
	return getSubscriptionsFunc()
}

// This mocks the NewClient function and returns an instance of the subscriptionGetterMock
type subscriptionClientGeneratorMock struct{}

func (m *subscriptionClientGeneratorMock) NewClient(config subscription.SubscriptionRequest) subscription.SubscriptionGetter {
	return &subscriptionGetter
}

// clustersResponse takes in a file with subscription data and returns a new mock function
func subscriptionResponse(testdata string) func() ([]subscription.Subscription, error) {
	return func() ([]subscription.Subscription, error) {
		file, _ := ioutil.ReadFile(testdata)
		subscriptions := []subscription.Subscription{}
		err := json.Unmarshal([]byte(file), &subscriptions)
		return subscriptions, err
	}
}

func TestDiscoverClusters(t *testing.T) {
	type args struct {
		token   string
		baseURL string
		filters discoveryv1.Filter
	}
	tests := []struct {
		name             string
		authfunc         func(auth.AuthRequest) (string, error)
		subscriptionFunc func() ([]subscription.Subscription, error)
		args             args
		want             int
		wantErr          bool
	}{
		{
			name: "Complete subscription",
			authfunc: func(auth.AuthRequest) (string, error) {
				// this mock return a dummy token
				return "valid_access_token", nil
			},
			// this mock returns 3 subscriptions read from mock_subscriptions.json
			subscriptionFunc: subscriptionResponse("testdata/1_mock_subscription.json"),
			args: args{
				token:   "test",
				baseURL: "test",
				filters: discoveryv1.Filter{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Two complete subscriptions, one incomplete",
			authfunc: func(auth.AuthRequest) (string, error) {
				// this mock return a dummy token
				return "valid_access_token", nil
			},
			// this mock returns 3 subscriptions read from mock_subscriptions.json
			subscriptionFunc: subscriptionResponse("testdata/3_mock_subscriptions.json"),
			args: args{
				token:   "test",
				baseURL: "test",
				filters: discoveryv1.Filter{},
			},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth.AuthClient = &authServiceMock{}                                          // Mocks out the call to auth service
			subscription.SubscriptionClientGenerator = &subscriptionClientGeneratorMock{} // Mocks out the subscription client creation

			getTokenFunc = tt.authfunc
			getSubscriptionsFunc = tt.subscriptionFunc

			got, err := DiscoverClusters(tt.args.token, tt.args.baseURL, tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverClusters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("DiscoverClusters() = %v, wanted %d clusters", got, tt.want)
			}
		})
	}
}