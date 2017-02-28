package oauth

import (
	"github.com/imdario/mergo"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/amazon"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/box"
	"github.com/markbates/goth/providers/digitalocean"
	"github.com/markbates/goth/providers/dropbox"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/gplus"
	"github.com/markbates/goth/providers/heroku"
	"github.com/markbates/goth/providers/instagram"
	"github.com/markbates/goth/providers/lastfm"
	"github.com/markbates/goth/providers/linkedin"
	"github.com/markbates/goth/providers/onedrive"
	"github.com/markbates/goth/providers/paypal"
	"github.com/markbates/goth/providers/salesforce"
	"github.com/markbates/goth/providers/slack"
	"github.com/markbates/goth/providers/soundcloud"
	"github.com/markbates/goth/providers/spotify"
	"github.com/markbates/goth/providers/steam"
	"github.com/markbates/goth/providers/stripe"
	"github.com/markbates/goth/providers/twitch"
	"github.com/markbates/goth/providers/twitter"
	"github.com/markbates/goth/providers/uber"
	"github.com/markbates/goth/providers/wepay"
	"github.com/markbates/goth/providers/yahoo"
	"github.com/markbates/goth/providers/yammer"
)

const (

	// DefaultRouteName /oauth
	//DefaultPath = "/oauth"
	// DefaultRouteName oauth
	DefaultRouteName = "oauth"
	// DefaultRequestParamPath is the default named path parameter for the url provider
	DefaultRequestParamPath = "provider"
	// DefaultContextKey oauth_user
	DefaultContextKey = "oauth_user"
)

// Config the configs for the gothic oauth/oauth2 authentication for third-party websites
// All Key and Secret values are empty by default strings. Non-empty will be registered as Goth Provider automatically, by Iris
// the users can still register their own providers using goth.UseProviders
// contains the providers' keys  (& secrets) and the relative auth callback url path(ex: "/auth" will be registered as /auth/:provider/callback)
//
type Config struct {
	// Author's notes:
	// after 6.1.4 we need these three fields because routers are setted by
	// adaptors and some works with : other with {}
	// and who knows how many of them the users will adapt to iris, so ask for explicit paths here.

	// RequestPath for example "/auth/{provider}"
	RequestPath      string
	RequestPathParam string // if RequestPath is: "/auth/{provider}" then this field should be "provider""
	// CallbackRelativePath relative to RequestPath, for example "/callback"
	// it will convert to: "/auth/facebook/callback" if facebook provider.
	CallbackRelativePath string

	TwitterKey, TwitterSecret, TwitterName                string
	FacebookKey, FacebookSecret, FacebookName             string
	GplusKey, GplusSecret, GplusName                      string
	GithubKey, GithubSecret, GithubName                   string
	SpotifyKey, SpotifySecret, SpotifyName                string
	LinkedinKey, LinkedinSecret, LinkedinName             string
	LastfmKey, LastfmSecret, LastfmName                   string
	TwitchKey, TwitchSecret, TwitchName                   string
	DropboxKey, DropboxSecret, DropboxName                string
	DigitaloceanKey, DigitaloceanSecret, DigitaloceanName string
	BitbucketKey, BitbucketSecret, BitbucketName          string
	InstagramKey, InstagramSecret, InstagramName          string
	BoxKey, BoxSecret, BoxName                            string
	SalesforceKey, SalesforceSecret, SalesforceName       string
	AmazonKey, AmazonSecret, AmazonName                   string
	YammerKey, YammerSecret, YammerName                   string
	OneDriveKey, OneDriveSecret, OneDriveName             string
	YahooKey, YahooSecret, YahooName                      string
	SlackKey, SlackSecret, SlackName                      string
	StripeKey, StripeSecret, StripeName                   string
	WepayKey, WepaySecret, WepayName                      string
	PaypalKey, PaypalSecret, PaypalName                   string
	SteamKey, SteamName                                   string
	HerokuKey, HerokuSecret, HerokuName                   string
	UberKey, UberSecret, UberName                         string
	SoundcloudKey, SoundcloudSecret, SoundcloudName       string
	GitlabKey, GitlabSecret, GitlabName                   string

	//RouteName is the registered route's name, using to help you render a link using templates or iris.URL("RouteName","Providername")
	// defaults to 'oauth'
	RouteName string
	// defaults to 'oauth_user' used by plugin to give you the goth.User, but you can take this manually also by `context.Get(ContextKey).(goth.User)`
	ContextKey string
}

// DefaultConfig returns OAuth config, the fields of the iteral are zero-values ( empty strings)
func DefaultConfig() Config {
	return Config{
		TwitterName:      "twitter",
		FacebookName:     "facebook",
		GplusName:        "gplus",
		GithubName:       "github",
		SpotifyName:      "spotify",
		LinkedinName:     "linkedin",
		LastfmName:       "lastfm",
		TwitchName:       "twitch",
		DropboxName:      "dropbox",
		DigitaloceanName: "digitalocean",
		BitbucketName:    "bitbucket",
		InstagramName:    "instagram",
		BoxName:          "box",
		SalesforceName:   "salesforce",
		AmazonName:       "amazon",
		YammerName:       "yammer",
		OneDriveName:     "onedrive",
		YahooName:        "yahoo",
		SlackName:        "slack",
		StripeName:       "stripe",
		WepayName:        "wepay",
		PaypalName:       "paypal",
		SteamName:        "steam",
		HerokuName:       "heroku",
		UberName:         "uber",
		SoundcloudName:   "soundcloud",
		GitlabName:       "gitlab",
		RouteName:        DefaultRouteName,
		ContextKey:       DefaultContextKey,
		RequestPathParam: DefaultRequestParamPath,
	}
}

// MergeSingle merges the default with the given config and returns the result
func (c Config) MergeSingle(cfg Config) (config Config) {

	config = cfg
	mergo.Merge(&config, c)
	return
}

// GenerateProviders returns the valid goth providers and the relative url paths (because the goth.Provider doesn't have a public method to get the Auth path...)
// we do the hard-core/hand checking here at the configs.
//
// receives one parameter which is the host from the server,ex: http://localhost:3000, will be used as prefix for the oauth callback
func (c Config) GenerateProviders(vhost string) (providers []goth.Provider) {
	getCallbackURL := func(providerName string) string {
		//println("Registered : "+ vhost + c.RequestPath + "/" + providerName + c.CallbackRelativePath)
		return vhost + c.RequestPath + "/" + providerName + c.CallbackRelativePath
	}

	//we could use a map but that's easier for the users because of code completion of their IDEs/editors

	//Contributor's note.
	//Next time I have to rewrite it... I make a map anyway... My hand hurts because of copy/paste...

	if c.TwitterKey != "" && c.TwitterSecret != "" {
		provider := twitter.New(c.TwitterKey, c.TwitterSecret, getCallbackURL(c.TwitterName))
		if c.TwitterName != "" {
			provider.SetName(c.TwitterName)
		}
		providers = append(providers, provider)
	}
	if c.FacebookKey != "" && c.FacebookSecret != "" {
		provider := facebook.New(c.FacebookKey, c.FacebookSecret, getCallbackURL(c.FacebookName))
		if c.FacebookName != "" {
			provider.SetName(c.FacebookName)
		}
		providers = append(providers, provider)
	}
	if c.GplusKey != "" && c.GplusSecret != "" {
		provider := gplus.New(c.GplusKey, c.GplusSecret, getCallbackURL(c.GplusName))
		if c.GplusName != "" {
			provider.SetName(c.GplusName)
		}
		providers = append(providers, provider)
	}
	if c.GithubKey != "" && c.GithubSecret != "" {
		provider := github.New(c.GithubKey, c.GithubSecret, getCallbackURL(c.GithubName))
		if c.GithubName != "" {
			provider.SetName(c.GithubName)
		}
		providers = append(providers, provider)
	}
	if c.SpotifyKey != "" && c.SpotifySecret != "" {
		provider := spotify.New(c.SpotifyKey, c.SpotifySecret, getCallbackURL(c.SpotifyName))
		if c.SpotifyName != "" {
			provider.SetName(c.SpotifyName)
		}
		providers = append(providers, provider)
	}
	if c.LinkedinKey != "" && c.LinkedinSecret != "" {
		provider := linkedin.New(c.LinkedinKey, c.LinkedinSecret, getCallbackURL(c.LinkedinName))
		if c.LinkedinName != "" {
			provider.SetName(c.LinkedinName)
		}
		providers = append(providers, provider)
	}
	if c.LastfmKey != "" && c.LastfmSecret != "" {
		provider := lastfm.New(c.LastfmKey, c.LastfmSecret, getCallbackURL(c.LastfmName))
		if c.LastfmName != "" {
			provider.SetName(c.LastfmName)
		}
		providers = append(providers, provider)
	}
	if c.TwitchKey != "" && c.TwitchSecret != "" {
		provider := twitch.New(c.TwitchKey, c.TwitchSecret, getCallbackURL(c.TwitchName))
		if c.TwitchName != "" {
			provider.SetName(c.TwitchName)
		}
		providers = append(providers, provider)
	}
	if c.DropboxKey != "" && c.DropboxSecret != "" {
		provider := dropbox.New(c.DropboxKey, c.DropboxSecret, getCallbackURL(c.DropboxName))
		if c.DropboxName != "" {
			provider.SetName(c.DropboxName)
		}
		providers = append(providers, provider)
	}
	if c.DigitaloceanKey != "" && c.DigitaloceanSecret != "" {
		provider := digitalocean.New(c.DigitaloceanKey, c.DigitaloceanSecret, getCallbackURL(c.DigitaloceanName))
		if c.DigitaloceanName != "" {
			provider.SetName(c.DigitaloceanName)
		}
		providers = append(providers, provider)
	}
	if c.BitbucketKey != "" && c.BitbucketSecret != "" {
		provider := bitbucket.New(c.BitbucketKey, c.BitbucketSecret, getCallbackURL(c.BitbucketName))
		if c.BitbucketName != "" {
			provider.SetName(c.BitbucketName)
		}
		providers = append(providers, provider)
	}
	if c.InstagramKey != "" && c.InstagramSecret != "" {
		provider := instagram.New(c.InstagramKey, c.InstagramSecret, getCallbackURL(c.InstagramName))
		if c.InstagramName != "" {
			provider.SetName(c.InstagramName)
		}
		providers = append(providers, provider)
	}
	if c.BoxKey != "" && c.BoxSecret != "" {
		provider := box.New(c.BoxKey, c.BoxSecret, getCallbackURL(c.BoxName))
		if c.BoxName != "" {
			provider.SetName(c.BoxName)
		}
		providers = append(providers, provider)
	}
	if c.SalesforceKey != "" && c.SalesforceSecret != "" {
		provider := salesforce.New(c.SalesforceKey, c.SalesforceSecret, getCallbackURL(c.SalesforceName))
		if c.SalesforceName != "" {
			provider.SetName(c.SalesforceName)
		}
		providers = append(providers, provider)
	}
	if c.AmazonKey != "" && c.AmazonSecret != "" {
		provider := amazon.New(c.AmazonKey, c.AmazonSecret, getCallbackURL(c.AmazonName))
		if c.AmazonName != "" {
			provider.SetName(c.AmazonName)
		}
		providers = append(providers, provider)
	}
	if c.YammerKey != "" && c.YammerSecret != "" {
		provider := yammer.New(c.YammerKey, c.YammerSecret, getCallbackURL(c.YammerName))
		if c.YammerName != "" {
			provider.SetName(c.YammerName)
		}
		providers = append(providers, provider)
	}
	if c.OneDriveKey != "" && c.OneDriveSecret != "" {
		provider := onedrive.New(c.OneDriveKey, c.OneDriveSecret, getCallbackURL(c.OneDriveName))
		if c.YammerName != "" {
			provider.SetName(c.OneDriveName)
		}
		providers = append(providers, provider)
	}
	if c.YahooKey != "" && c.YahooSecret != "" {
		provider := yahoo.New(c.YahooKey, c.YahooSecret, getCallbackURL(c.YahooName))
		if c.YahooName != "" {
			provider.SetName(c.YahooName)
		}
		providers = append(providers, provider)
	}
	if c.SlackKey != "" && c.SlackSecret != "" {
		provider := slack.New(c.SlackKey, c.SlackSecret, getCallbackURL(c.SlackName))
		if c.SlackName != "" {
			provider.SetName(c.SlackName)
		}
		providers = append(providers, provider)
	}
	if c.StripeKey != "" && c.StripeSecret != "" {
		provider := stripe.New(c.StripeKey, c.StripeSecret, getCallbackURL(c.StripeName))
		if c.StripeName != "" {
			provider.SetName(c.StripeName)
		}
		providers = append(providers, provider)
	}
	if c.WepayKey != "" && c.WepaySecret != "" {
		provider := wepay.New(c.WepayKey, c.WepaySecret, getCallbackURL(c.WepayName))
		if c.WepayName != "" {
			provider.SetName(c.WepayName)
		}
		providers = append(providers, provider)
	}
	if c.PaypalKey != "" && c.PaypalSecret != "" {
		provider := paypal.New(c.PaypalKey, c.PaypalSecret, getCallbackURL(c.PaypalName))
		if c.PaypalName != "" {
			provider.SetName(c.PaypalName)
		}
		providers = append(providers, provider)
	}
	if c.SteamKey != "" {
		provider := steam.New(c.SteamKey, getCallbackURL(c.SteamName))
		if c.SteamName != "" {
			provider.SetName(c.SteamName)
		}
		providers = append(providers, provider)
	}
	if c.HerokuKey != "" && c.HerokuSecret != "" {
		provider := heroku.New(c.HerokuKey, c.HerokuSecret, getCallbackURL(c.HerokuName))
		if c.HerokuName != "" {
			provider.SetName(c.HerokuName)
		}
		providers = append(providers, provider)
	}
	if c.UberKey != "" && c.UberSecret != "" {
		provider := uber.New(c.UberKey, c.UberSecret, getCallbackURL(c.UberName))
		if c.UberName != "" {
			provider.SetName(c.UberName)
		}
		providers = append(providers, provider)
	}
	if c.SoundcloudKey != "" && c.SoundcloudSecret != "" {
		provider := soundcloud.New(c.SoundcloudKey, c.SoundcloudSecret, getCallbackURL(c.SoundcloudName))
		if c.SoundcloudName != "" {
			provider.SetName(c.SoundcloudName)
		}
		providers = append(providers, provider)
	}
	if c.GitlabKey != "" && c.GitlabSecret != "" {
		provider := gitlab.New(c.GitlabKey, c.GitlabSecret, getCallbackURL(c.GitlabName))
		if c.GitlabName != "" {
			provider.SetName(c.GitlabName)
		}
		providers = append(providers, provider)
	}

	return
}
