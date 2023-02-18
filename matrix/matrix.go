package matrix

// #cgo CFLAGS: -O3 -march=native
// #include "matrix.h"
import "C"

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	mrand "math/rand"
  "reflect"

  "github.com/henrycg/simplepir/lwe"
)

type Elem32 = C.Elem32
type Elem64 = C.Elem64

type Elem interface {
    Elem32 | Elem64
}

type IoRandSource interface {
	io.Reader
	mrand.Source64
}

type Matrix[T Elem] struct {
	rows uint64
	cols uint64
	data []T
}

func (m *Matrix[T]) Is32Bit() bool {
  return reflect.TypeOf(m.data[0]) == reflect.TypeOf(Elem32(0))
}

func (m *Matrix[T]) Is64Bit() bool {
  return reflect.TypeOf(m.data[0]) == reflect.TypeOf(Elem64(0))
}

func (m *Matrix[T]) Copy() *Matrix[T] {
	out := &Matrix[T]{
		rows: m.rows,
		cols: m.cols,
		data: make([]T, len(m.data)),
	}

	copy(out.data[:], m.data[:])
	return out
}

func (m *Matrix[T]) Rows() uint64 {
	return m.rows
}

func (m *Matrix[T]) Cols() uint64 {
	return m.cols
}

func (m *Matrix[T]) Size() uint64 {
	return m.rows * m.cols
}

func (m *Matrix[T]) AppendZeros(n uint64) {
	m.Concat(Zeros[T](n, 1))
}

func New[T Elem](rows uint64, cols uint64) *Matrix[T] {
	out := new(Matrix[T])
	out.rows = rows
	out.cols = cols
	out.data = make([]T, rows*cols)
	return out
}

func Rand[T Elem](src IoRandSource, rows uint64, cols uint64, logmod uint64, mod uint64) *Matrix[T] {
	out := New[T](rows, cols)
	m := big.NewInt(int64(mod))
	if mod == 0 {
		m = big.NewInt(1 << logmod)
	}
	for i := 0; i < len(out.data); i++ {
		v, err := rand.Int(src, m)
		if err != nil {
			panic("Randomness error")
		}
		out.data[i] = T(v.Uint64())
	}
	return out
}

func Zeros[T Elem](rows uint64, cols uint64) *Matrix[T] {
	out := New[T](rows, cols)
	for i := 0; i < len(out.data); i++ {
		out.data[i] = T(0)
	}
	return out
}

func (m *Matrix[T]) ReduceMod(p uint64) {
	mod := T(p)
	for i := 0; i < len(m.data); i++ {
		m.data[i] = m.data[i] % mod
	}
}

func (m *Matrix[T]) Get(i, j uint64) uint64 {
	if i >= m.rows {
		panic("Too many rows!")
	}
	if j >= m.cols {
		panic("Too many cols!")
	}
	return uint64(m.data[i*m.cols+j])
}

func (m *Matrix[T]) Set(val, i, j uint64) {
	if i >= m.rows {
		panic("Too many rows!")
	}
	if j >= m.cols {
		panic("Too many cols!")
	}
	m.data[i*m.cols+j] = T(val)
}


func (a *Matrix[T]) Concat(b *Matrix[T]) {
	if a.cols == 0 && a.rows == 0 {
		a.cols = b.cols
		a.rows = b.rows
		a.data = b.data
		return
	}

	if a.cols != b.cols {
		fmt.Printf("%d-by-%d vs. %d-by-%d\n", a.rows, a.cols, b.rows, b.cols)
		panic("Dimension mismatch")
	}

	a.rows += b.rows
	a.data = append(a.data, b.data...)
}

func (m *Matrix[T]) DropLastrows(n uint64) {
	m.rows -= n
	m.data = m.data[:(m.rows * m.cols)]
}

func (m *Matrix[T]) GetRow(offset, num_rows uint64) *Matrix[T] {
	if offset+num_rows > m.rows {
		panic("Requesting too many rows")
	}

	m2 := New[T](num_rows, m.cols)
	m2.data = m.data[(offset * m.cols):((offset + num_rows) * m.cols)]
	return m2
}

func (m *Matrix[T]) rowsDeepCopy(offset, num_rows uint64) *Matrix[T] {
	if offset+num_rows > m.rows {
		panic("Requesting too many rows")
	}

	if offset+num_rows <= m.rows {
		m2 := New[T](num_rows, m.cols)
		copy(m2.data, m.data[(offset*m.cols):((offset+num_rows)*m.cols)])
		return m2
	}

	m2 := New[T](m.rows-offset, m.cols)
	copy(m2.data, m.data[(offset*m.cols):(m.rows)*m.cols])
	return m2
}

func (m *Matrix[T]) Dim() {
	fmt.Printf("Dims: %d-by-%d\n", m.rows, m.cols)
}

func (m *Matrix[T]) Equals(n *Matrix[T]) bool {
	if m.Cols() != n.Cols() {
		return false
	}
	if m.Rows() != n.Rows() {
		return false
	}
	return reflect.DeepEqual(m.data, n.data)
}


func Gaussian[T Elem](src IoRandSource, rows, cols uint64) *Matrix[T] {
	out := New[T](rows, cols)
  samplef := lwe.GaussSample32
  if out.Is32Bit() {
      // Do nothing
  } else if out.Is64Bit() {
    samplef = lwe.GaussSample64
  } else {
    panic("Shouldn't get here")
  }

	for i := 0; i < len(out.data); i++ {
		out.data[i] = T(samplef(src))
	}
	return out
}

