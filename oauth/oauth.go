package oauth

import (
	"github.com/kataras/go-errors"
	"github.com/markbates/goth"
	"gopkg.in/kataras/iris.v6"
	"strings"
)

// SessionValueKey is the key used to access the session store.
var SessionValueKey = "auth.session"

// Oauth is an adaptor which helps you to use OAuth/OAuth2 apis from famous websites
type Oauth struct {
	Config          Config
	successHandlers []iris.HandlerFunc
	failHandler     iris.HandlerFunc
	station         *iris.Framework
	middleware      iris.HandlerFunc
}

// New returns a new OAuth Oauth
// receives one parameter of type 'Config'
func New(cfg Config) *Oauth {
	c := DefaultConfig().MergeSingle(cfg)
	//Set default middleware
	middleware := func(c *iris.Context) { c.Next() }
	return &Oauth{Config: c, middleware: middleware}
}

// Success registers handler(s) which fires when the user logged in successfully
func (p *Oauth) Success(handlersFn ...iris.HandlerFunc) {
	p.successHandlers = append(p.successHandlers, handlersFn...)
}

// Fail registers handler which fires when the user failed to logged in
// underhood it justs registers an error handler to the StatusUnauthorized(400 status code), same as 'iris.Default.OnError(400,handler)'
func (p *Oauth) Fail(handler iris.HandlerFunc) {
	p.failHandler = handler
}

//Use registers a handler which fires before the user logs in
func (p *Oauth) Use(handler iris.HandlerFunc) {
	p.middleware = handler
}

// User returns the user for the particular client
// if user is not validated  or not found it returns nil
// same as 'ctx.Get(config's ContextKey field).(User)'
func (p *Oauth) User(ctx *iris.Context) (u User) {
	return ctx.Get(p.Config.ContextKey).(User)
}

// URL returns the full URL of a provider
// Use this method to get the url which you will render on your html page to create a link for user authentication
//
// same as `iris.URL(config's RouteName field, "theprovidername")`
// notes:
// If you use the Iris' view system then you can use the {{url }} func inside your template directly:
// {{ url config's RouteName field, "theprovidername"}} |  example: {{url "oauth" "facebook"}}, "oauth" is ,also, the route's name , so this will give the http(s)://yourhost:port/oauth/facebook
func (p *Oauth) URL(providerName string) string {
	return p.station.URL(p.Config.RouteName, providerName)
}

// Adapt adapts the oauth2 adaptor.
// Note:
// We use that method and not the return on New because we
// want to export the Oauth's functionality to the user.
func (p *Oauth) Adapt(frame *iris.Policies) {
	policy := iris.EventPolicy{
		Boot: p.boot,
	}

	policy.Adapt(frame)
}

// boot builds the Oauth in order to be registered to the iris
// boot because we add routes.
func (p *Oauth) boot(s *iris.Framework) {
	if p.Config.RequestPath == "" || p.Config.CallbackRelativePath == "" || p.Config.RequestPathParam == "" {
		s.Log(iris.ProdMode, "oauth adaptor disabled: Config.RequestPath or/and RequestPathParam or/and Config.CallbackRelativePath are empty,\nplease set them and restart the app")
		return
	}
	var oauthProviders []goth.Provider
	if p.Config.FQDN==""{
		oauthProviders = p.Config.GenerateProviders(s.Config.VScheme + s.Config.VHost)
	}else{
		VHost := strings.Split(s.Config.VHost,":")
		oauthProviders = p.Config.GenerateProviders(s.Config.VScheme + p.Config.FQDN + VHost[1])
	}
	if len(oauthProviders) > 0 {
		goth.UseProviders(oauthProviders...)
		// set the mux path to handle the registered providers

		//Expected behavior
		//http://domain.com  /ROUTE /  {provider}
		//     VHOST        REQ PATH    PARAM
		//println("Config is : requPath = " +p.Config.RequestPath)
		s.Get(p.Config.RequestPath+"/"+s.RouteParam(p.Config.RequestPathParam), p.middleware, func(ctx *iris.Context) {
			err := p.BeginAuthHandler(ctx)
			if err != nil {
				s.Log(iris.DevMode, "oauth adaptor runtime error on '"+ctx.Path()+"'. Trace: "+err.Error())
			}
		}).ChangeName(p.Config.RouteName)

		//println("registered " + p.Config.RequestPath+"/{"+p.Config.RequestPathParam+"}")

		authMiddleware := func(ctx *iris.Context) {
			user, err := p.CompleteUserAuth(ctx)
			if err != nil {
				ctx.EmitError(iris.StatusUnauthorized)
				s.Log(iris.DevMode, "oauth adaptor runtime error on '"+ctx.Path()+"'. Trace: "+err.Error())
				return
			}
			ctx.Set(p.Config.ContextKey, user)
			ctx.Next()
		}

		p.successHandlers = append([]iris.HandlerFunc{authMiddleware}, p.successHandlers...)

		//Expected behavior
		//http://domain.com  /ROUTE   {provider}      /callback
		//     VHOST        REQ PATH    PARAM     CALLBACKRELATIVE
		//println("Config is : requPath = " +p.Config.RequestPath)
		//println("param name = "+p.Config.RequestPathParam)
		//println("callback relative = "+ p.Config.CallbackRelativePath)

		s.Get(p.Config.RequestPath+"/"+s.RouteParam(p.Config.RequestPathParam)+p.Config.CallbackRelativePath, p.successHandlers...)
		p.station = s
		// register the error handler
		if p.failHandler != nil {
			s.OnError(iris.StatusUnauthorized, p.failHandler)
		}
	}
}

// BeginAuthHandler is a convienence handler for starting the authentication process.
// It expects to be able to get the name of the provider from the named parameters
// as either "provider" or url query parameter ":provider".

// BeginAuthHandler will redirect the user to the appropriate authentication end-point
// for the requested provider.
func (p *Oauth) BeginAuthHandler(ctx *iris.Context) error {
	url, err := p.GetAuthURL(ctx)
	if err != nil {
		ctx.NotFound()
		return err
	}

	ctx.Redirect(url)
	return nil
}

// GetAuthURL starts the authentication process with the requested provided.
// It will return a URL that should be used to send users to.
//
// It expects to be able to get the name of the provider from the query parameters
// as either "provider" or url query parameter ":provider".
//
// I would recommend using the BeginAuthHandler instead of doing all of these steps
// yourself, but that's entirely up to you.
func (p *Oauth) GetAuthURL(ctx *iris.Context) (string, error) {

	providerName, err := p.GetProviderName(ctx)
	if err != nil {
		return "", err
	}
	//println("Provider name is : "+providerName)
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	sess, err := provider.BeginAuth(setState(ctx))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	ctx.Session().Set(SessionValueKey, sess.Marshal())

	return url, nil
}

// User is a conversion for the User
type User goth.User

// CompleteUserAuth does what it says on the tin. It completes the authentication
// process and fetches all of the basic information about the user from the provider.
//
// It expects to be able to get the name of the provider from the named parameters
// as either "provider" or url query parameter "provider".
func (p *Oauth) CompleteUserAuth(ctx *iris.Context) (User, error) {

	providerName, err := p.GetProviderName(ctx)
	if err != nil {
		return User{}, err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return User{}, err
	}

	if ctx.Session().Get(SessionValueKey) == nil {
		return User{}, errors.New("completeUserAuth error: could not find a matching session for this request")
	}

	sess, err := provider.UnmarshalSession(ctx.Session().GetString(SessionValueKey))
	if err != nil {
		return User{}, err
	}
	_, err = sess.Authorize(provider, ctx.Request.URL.Query())

	if err != nil {
		return User{}, err
	}
	u, err := provider.FetchUser(sess)
	return User(u), err
}

// GetProviderName is a function used to get the name of a provider
// for a given request.This provider is fetched from
// the URL query string or named parameter (p.Config.RequestPathParam).
func (p *Oauth) GetProviderName(ctx *iris.Context) (string, error) {
	provider := ctx.Param(p.Config.RequestPathParam)
	if provider == "" {
		provider = ctx.URLParam(p.Config.RequestPathParam)
	}
	if provider == "" {
		return provider, errors.New("getProviderName error: you must select a provider")
	}
	return provider, nil
}

// SetState sets the state string associated with the given request.
// If no state string is associated with the request, one will be generated.
// This state is sent to the provider and can be retrieved during the
// callback.
func setState(ctx *iris.Context) string {
	state := ctx.URLParam("state")
	if len(state) > 0 {
		return state
	}

	return "state"

}

// GetState gets the state returned by the provider during the callback.
// This is used to prevent CSRF attacks, see
// http://tools.ietf.org/html/rfc6749#section-10.12
func GetState(ctx *iris.Context) string {
	return ctx.URLParam("state")
}
