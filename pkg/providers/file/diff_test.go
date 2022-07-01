package file

import (
	"github.com/michaellee8/dynproxy/pkg/proxy/op"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiffConfig(t *testing.T) {

	t.Run("Simple", func(t *testing.T) {
		expectedOps := []op.Operation{
			op.RemoveRuleOperation("abc"),
			op.AddRuleOperation("ddd"),
		}

		oldCfg := Config{
			Rules: RuleSlice{
				{
					Key:     "abc",
					Targets: nil,
					Ports:   nil,
				},
				{
					Key:     "efg",
					Targets: nil,
					Ports:   nil,
				},
			},
		}

		newCfg := Config{
			Rules: RuleSlice{
				{
					Key:     "efg",
					Targets: nil,
					Ports:   nil,
				},
				{
					Key:     "ddd",
					Targets: nil,
					Ports:   nil,
				},
			},
		}

		assert.Equal(t, expectedOps, DiffConfig(oldCfg, newCfg))
	})

	t.Run("Complex", func(t *testing.T) {
		oldCfg := Config{
			Rules: RuleSlice{
				{
					Key: "seven-thousand",
					Ports: []int{
						7300,
						7400,
						8081,
						8082,
					},
					Targets: []string{
						"localhost:7002",
						"localhost:8081",
					},
				},
				{
					Key: "zzz",
					Ports: []int{
						1234,
						5678,
					},
					Targets: []string{
						"localhost:5002",
						"localhost:5001",
					},
				},
				{
					Key: "five-thousand",
					Ports: []int{
						5400,
						5200,
						10667,
						10666,
					},
					Targets: []string{
						"localhost:10001",
						"localhost:10000",
						"localhost:5001",
					},
				},
			},
		}

		newCfg := Config{
			Rules: RuleSlice{
				{
					Key: "seven-thousand",
					Ports: []int{
						7001,
						7200,
						7300,
						7400,
					},
					Targets: []string{
						"localhost:7001",
						"localhost:7002",
					},
				},
				{
					Key: "six-thousand",
					Ports: []int{
						6001,
						6200,
						6300,
						6400,
					},
					Targets: []string{
						"localhost:6001",
						"localhost:6002",
						"bad-target.for-testing:6003",
					},
				},
				{
					Key: "five-thousand",
					Ports: []int{
						5001,
						5200,
						5300,
						5400,
					},
					Targets: []string{
						"localhost:5001",
						"localhost:5002",
					},
				},
			},
		}

		expectedOps := []op.Operation{
			op.RemovePortOperation("five-thousand", 10666),
			op.RemovePortOperation("five-thousand", 10667),
			op.RemovePortOperation("seven-thousand", 8081),
			op.RemovePortOperation("seven-thousand", 8082),
			op.RemoveTargetOperation("five-thousand", "localhost:10000"),
			op.RemoveTargetOperation("five-thousand", "localhost:10001"),
			op.RemoveTargetOperation("seven-thousand", "localhost:8081"),
			op.RemoveRuleOperation("zzz"),
			op.AddRuleOperation("six-thousand"),
			op.AddTargetOperation("six-thousand", "bad-target.for-testing:6003"),
			op.AddTargetOperation("six-thousand", "localhost:6001"),
			op.AddTargetOperation("six-thousand", "localhost:6002"),
			op.AddTargetOperation("five-thousand", "localhost:5002"),
			op.AddTargetOperation("seven-thousand", "localhost:7001"),
			op.AddPortOperation("six-thousand", 6001),
			op.AddPortOperation("six-thousand", 6200),
			op.AddPortOperation("six-thousand", 6300),
			op.AddPortOperation("six-thousand", 6400),
			op.AddPortOperation("five-thousand", 5001),
			op.AddPortOperation("five-thousand", 5300),
			op.AddPortOperation("seven-thousand", 7001),
			op.AddPortOperation("seven-thousand", 7200),
		}

		actualOps := DiffConfig(oldCfg, newCfg)
		assert.Equal(t, expectedOps, actualOps)
	})

}
