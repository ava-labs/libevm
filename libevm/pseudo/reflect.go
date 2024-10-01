// Copyright 2024 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.

package pseudo

import "reflect"

// Reflection is used as a last resort in pseudo types so is limited to this
// file to avoid being seen as the norm. If you are adding to this file, please
// try to achieve the same results with type parameters.

func (c *concrete[T]) ensureNonNilPointer() {
	v := reflect.ValueOf(c.val)
	if v.Kind() != reflect.Pointer || !v.IsNil() {
		return
	}
	el := v.Type().Elem()
	c.val = reflect.New(el).Interface().(T) //nolint:forcetypeassert // Invariant scoped to the last few lines of code so simple to verify
}
