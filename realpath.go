// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package realpath

import (
	"bytes"
	"fmt"
	"os"
)

func Realpath(filepath string) (string, error) {

	if len(filepath) == 0 {
		return "", os.ErrInvalid
	}

	sepStr := string(os.PathSeparator)

	if filepath[0] != os.PathSeparator {
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		filepath = pwd + sepStr + filepath
	}

	path := []byte(filepath)
	nlinks := 0
	start := 1
	prev := 1
	for start < len(path) {
		c := nextComponent(path, start)
		cur := c[start:]

		switch {

		case len(cur) == 0:
			copy(path[start:], path[start+1:])
			path = path[0 : len(path)-1]

		case len(cur) == 1 && cur[0] == '.':
			if start+2 < len(path) {
				copy(path[start:], path[start+2:])
			}
			path = path[0 : len(path)-2]

		case len(cur) == 2 && cur[0] == '.' && cur[1] == '.':
			copy(path[prev:], path[start+2:])
			path = path[0 : len(path)+prev-(start+2)]
			prev = 1
			start = 1

		default:

			fi, err := os.Lstat(string(c))
			if err != nil {
				return "", err
			}
			if fi.Mode()&os.ModeSymlink == os.ModeSymlink {

				nlinks++
				if nlinks > 16 {
					return "", os.ErrInvalid
				}

				var dst string
				dst, err = os.Readlink(string(c))
				fmt.Printf("SYMLINK -> %s\n", dst)

				rest := string(path[len(c):])
				if dst[0] == os.PathSeparator {
					// Absolute links
					path = []byte(dst + sepStr + rest)
				} else {
					// Relative links
					path = []byte(string(path[0:start]) + dst + sepStr + rest)
				}
				prev = 1
				start = 1
			} else {
				// Directories
				prev = start
				start = len(c) + 1
			}
		}
	}

	for len(path) > 1 && path[len(path)-1] == os.PathSeparator {
		path = path[0 : len(path)-1]
	}
	return string(path), nil

}

func nextComponent(path []byte, start int) []byte {
	v := bytes.IndexByte(path[start:], os.PathSeparator)
	if v < 0 {
		return path
	}
	return path[0 : start+v]
}
