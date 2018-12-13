// +build go1.8

/*
================================================================
=  Source code from https://github.com/dominikh/go-tools       =
=  Copyright @ Dominik Honnef (https://github.com/dominikh)    =
================================================================
*/

package ssa

import "go/types"

var structTypesIdentical = types.IdenticalIgnoreTags
