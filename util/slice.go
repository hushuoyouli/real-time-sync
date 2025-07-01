package util

func AddToSlice[T interface{}, S []T](slice *S, elem T) {
	*slice = append(*slice, elem)
}

func RemoveFromSlice[T interface{}, S []T](slice *S, index int) {
	if index >= len(*slice) {
		return
	}

	if index < 0 {
		return
	}

	if index != len(*slice)-1 {
		copy((*slice)[index:len(*slice)-1], (*slice)[index+1:len(*slice)])
	}

	*slice = (*slice)[0 : len(*slice)-1]
}
