package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestIPv6Binding(t *testing.T) {
	// Let's create a random address
	orig := Address5D{
		X: rand.Float64(),
		Y: rand.Float64(),
		Z: rand.Float64(),
		T: rand.Float64(),
		W: rand.Float64(),
	}

	ip := orig.ToIPv6()
	restored := Address5DFromIPv6(ip)
	
	fmt.Printf("Orig:     %+v\n", orig)
	fmt.Printf("Restored: %+v\n", restored)
	fmt.Printf("IPv6:     %s\n", ip.String())
	
	b27 := Base27Encode(ip)
	fmt.Printf("Base27:   %s\n", b27)
	
	rip, err := Base27Decode(b27)
	if err != nil {
		t.Fatalf("Decode err: %v", err)
	}
	fmt.Printf("R-IPv6:   %s\n", rip.String())
}
