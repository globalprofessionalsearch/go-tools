# Go Tools #

[![Build Status](https://travis-ci.org/globalprofessionalsearch/go-tools.svg?branch=master)](https://travis-ci.org/globalprofessionalsearch/go-tools)

A collection of small packages bundled independently for reuse between projects.  See individual packages for more details.

As these packages are reused between projects, you should take care to separate packages by their required dependencies.  A project that wants to use one or two packages here shouldn't be forced to import 3rd party dependencies it does not need.  This is why there is a separation in the `auth` subpackages, for example.
