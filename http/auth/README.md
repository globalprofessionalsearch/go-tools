# Auth #

This package provides tools you can use in your app for handling some authentication and authorization functionality, via basic Api Key and JWT mechanisms.  There is no magic provided - it is up to your application to configure, register and use the functions properly.

## Design ##

When a request comes in, authentication middlewares check for an api key, or a JWT token.  If found, an *authenticator* is called to create an object, which is stored in the request context.  The object is defined by your application, and there are some small interfaces it can implement in order to work with the *authorizers* provided by this package.  The authenticators should be provided by your application - no assumptions are made about how your app should store or validate the credentials.  That said, in most cases your application will probably want to make database queries and return a custom struct that describes the clients in your application.

Next, route *authorization* middlewares can be used when setting up your routing to protect your http handlers, either by requiring that a `Client` be set in the request context, or by requiring that an `Authorizer` has the specified permissions.

As long as your app can create functions to validate incoming api keys, and/or JWT tokens, and return a custom struct that implements the `Client` and `Authorizer` interfaces, you can make use of the *authorization* middlewares provided.

See the tests for basic usage examples.

## Usage Example with Negroni & Gorilla ##

See the `auth_test` package tests for example usage.