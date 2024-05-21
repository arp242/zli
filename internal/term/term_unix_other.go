// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || linux || solaris || zos

package term

import "syscall"

const ioctlReadTermios = syscall.TCGETS
const ioctlWriteTermios = syscall.TCSETS