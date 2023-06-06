/*
 * Copyright (c) 2023 Gilles Chehade <gilles@poolp.org>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package sets

import "sync"

type Set struct {
	items   map[string]struct{}
	muItems sync.Mutex
}

func NewSet(items ...string) *Set {
	itemsMap := make(map[string]struct{})
	for _, item := range items {
		itemsMap[item] = struct{}{}
	}
	return &Set{
		items: itemsMap,
	}
}

func Union(sets ...*Set) *Set {
	Set := NewSet()
	for _, set := range sets {
		items := set.items
		for item := range items {
			Set.items[item] = struct{}{}
		}
	}
	return Set
}

func Intersection(sets ...*Set) *Set {
	intersectionMap := make(map[string]int)
	for _, set := range sets {
		items := set.items
		for item := range items {
			intersectionMap[item] += 1
		}
	}

	Set := NewSet()
	for k := range intersectionMap {
		if intersectionMap[k] == len(sets) {
			Set.items[k] = struct{}{}
		}
	}
	return Set
}

func Difference(sets ...*Set) *Set {
	sourceSet := sets[0]
	Set := NewSet()
	for item := range sourceSet.items {
		Set.items[item] = struct{}{}
	}
	for _, set := range sets[1:] {
		items := set.items
		for item := range items {
			delete(Set.items, item)
		}
	}
	return Set
}

func SymmetricDifference(sets ...*Set) *Set {
	intersectionMap := make(map[string]int)
	for _, set := range sets {
		for item := range set.items {
			intersectionMap[item] += 1
		}
	}

	Set := NewSet()
	for k := range intersectionMap {
		if intersectionMap[k] == 1 {
			Set.items[k] = struct{}{}
		}
	}
	return Set
}

func (s *Set) Items() map[string]struct{} {
	return s.items
}

func (s *Set) ItemsList() []string {
	s.muItems.Lock()
	defer s.muItems.Unlock()

	items := make([]string, 0)
	for item := range s.items {
		items = append(items, item)
	}
	return items
}

func (s *Set) Length() int64 {
	s.muItems.Lock()
	defer s.muItems.Unlock()
	return int64(len(s.items))
}

func (s *Set) Union(sets ...*Set) *Set {
	params := make([]*Set, 0)
	params = append(params, s)
	params = append(params, sets...)
	return Union(params...)
}

func (s *Set) Intersection(sets ...*Set) *Set {
	params := make([]*Set, 0)
	params = append(params, s)
	params = append(params, sets...)
	return Intersection(params...)
}

func (s *Set) Difference(sets ...*Set) *Set {
	params := make([]*Set, 0)
	params = append(params, s)
	params = append(params, sets...)
	return Difference(params...)
}

func (s *Set) SymmetricDifference(sets ...*Set) *Set {
	params := make([]*Set, 0)
	params = append(params, s)
	params = append(params, sets...)
	return SymmetricDifference(params...)
}

func (s *Set) Contains(value string) bool {
	if _, exists := s.items[value]; !exists {
		return false
	} else {
		return true
	}
}

func (s *Set) Add(value string) bool {
	s.muItems.Lock()
	defer s.muItems.Unlock()

	if _, exists := s.items[value]; !exists {
		s.items[value] = struct{}{}
		return true
	} else {
		return false
	}
}

func (s *Set) Remove(value string) bool {
	s.muItems.Lock()
	defer s.muItems.Unlock()

	if _, exists := s.items[value]; exists {
		delete(s.items, value)
		return true
	} else {
		return false
	}
}

func (s *Set) SupersetOf(target *Set) bool {
	if len(s.items) <= len(target.items) {
		return false
	}

	sourceMap := make(map[string]struct{})
	for item := range s.items {
		sourceMap[item] = struct{}{}
	}

	for item := range target.items {
		if _, exists := sourceMap[item]; !exists {
			return false
		}
	}

	return true
}

func (s *Set) SubsetOf(target *Set) bool {
	return target.SupersetOf(s)
}

func (s *Set) DisjointOf(target *Set) bool {
	sourceMap := make(map[string]struct{})
	for item := range s.items {
		sourceMap[item] = struct{}{}
	}

	targetMap := make(map[string]struct{})
	for item := range target.items {
		targetMap[item] = struct{}{}
	}

	if len(s.items) < len(target.items) {
		for item := range s.items {
			if _, exists := targetMap[item]; exists {
				return false
			}
		}
	} else {
		for item := range target.items {
			if _, exists := sourceMap[item]; exists {
				return false
			}
		}
	}
	return true
}

func (s *Set) SameAs(target *Set) bool {
	if len(s.items) != len(target.items) {
		return false
	}

	sourceMap := make(map[string]struct{})
	for item := range s.items {
		sourceMap[item] = struct{}{}
	}

	targetMap := make(map[string]struct{})
	for item := range target.items {
		targetMap[item] = struct{}{}
	}

	for item := range target.items {
		if _, exists := sourceMap[item]; !exists {
			return false
		}
	}

	return true
}
