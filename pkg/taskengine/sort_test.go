package taskengine

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSortTasksByLessWorkers(t *testing.T) {

	type testCase struct {
		input WorkerTasks
		want  WorkerTasks
	}

	// testCaseTasks from Ids
	T := func(tids ...string) Tasks {
		res := Tasks{}
		for _, tid := range tids {
			res = append(res, &testCaseTask{tid, 10, true})
		}
		return res
	}

	ts := T("t1", "t2", "t3", "t4", "t5")

	testCases := map[string]testCase{
		"different # of T for each W": {
			input: WorkerTasks{
				"w1": T("t3", "t2", "t1"),
				"w2": T("t3", "t2"),
				"w3": T("t3"),
			},
			want: WorkerTasks{
				"w1": T("t1", "t2", "t3"),
				"w2": T("t2", "t3"),
				"w3": T("t3"),
			},
		},
		"same T for all W": {
			input: WorkerTasks{
				"w1": T("t1", "t2", "t3"),
				"w2": T("t1", "t2", "t3"),
				"w3": T("t1", "t2", "t3"),
			},
			want: WorkerTasks{
				"w1": T("t1", "t2", "t3"),
				"w2": T("t2", "t3", "t1"),
				"w3": T("t3", "t1", "t2"),
			},
		},
		"almost same T for all W": {
			input: WorkerTasks{
				"w1": T("t1", "t2"),
				"w2": T("t1", "t2", "t3"),
				"w3": T("t1", "t2", "t3"),
			},
			want: WorkerTasks{
				"w1": T("t1", "t2"),
				"w2": T("t2", "t3", "t1"),
				"w3": T("t3", "t1", "t2"),
			},
		},
		"test case 3": {
			input: WorkerTasks{
				"w1": T("t3", "t2"),
				"w2": T("t3", "t1"),
				"w3": T("t1", "t2"),
			},
			want: WorkerTasks{
				"w1": T("t2", "t3"),
				"w2": T("t3", "t1"),
				"w3": T("t1", "t2"),
			},
		},
		"test case 4": {
			input: WorkerTasks{
				"w1": T("t1", "t2"),
				"w2": T("t1", "t2"),
				"w3": T("t1", "t2"),
			},
			want: WorkerTasks{
				"w1": T("t1", "t2"),
				"w2": T("t2", "t1"),
				"w3": T("t1", "t2"),
			},
		},
		"test case 5": {
			input: WorkerTasks{
				"w1": T("t2", "t1", "t7", "t8", "t9"),
				"w2": T("t4", "t3", "t7", "t8", "t9"),
				"w3": T("t6", "t5", "t7", "t8", "t9"),
			},
			want: WorkerTasks{
				"w1": T("t1", "t2", "t7", "t8", "t9"),
				"w2": T("t3", "t4", "t8", "t9", "t7"),
				"w3": T("t5", "t6", "t9", "t7", "t8"),
			},
		},
		"same tasks list object for all workers": {
			input: WorkerTasks{
				"w1": ts,
				"w2": ts,
				"w3": ts,
			},
			want: WorkerTasks{
				"w1": T("t1", "t2", "t3", "t4", "t5"),
				"w2": T("t2", "t3", "t4", "t5", "t1"),
				"w3": T("t3", "t4", "t5", "t1", "t2"),
			},
		},
	}

	copts := cmp.Options{
		//cmpopts.IgnoreUnexported(testCaseTask{}),
	}

	for title, tc := range testCases {
		got := tc.input
		got.SortTasks()

		if diff := cmp.Diff(tc.want, got, copts); diff != "" {
			t.Errorf("%s: mismatch (-want +got):\n%s", title, diff)
		}
	}
}

// {
// 	cryptonatorcom-EUR : [BTC, ETH]
// 	fondidocit : [LU0224105808, LU0244991385, LU0261960354, LU0267388220, LU0329206832, LU0408877842, LU0500207542, LU0846585023, LU1345485095]
// 	fundsquarenet : [LU0244991385, LU0261960354, LU0267388220, LU0329206832, LU0408877842, LU0500207542, LU0846585023, LU1345485095, LU0224105808]
// 	morningstarit : [LU0261960354, LU0267388220, LU0329206832, LU0408877842, LU0500207542, LU0846585023, LU1345485095, LU0224105808, LU0244991385]
//  }

func TestSort2(t *testing.T) {

	type testCase struct {
		input WorkerTasks
		want  WorkerTasks
	}

	// testCaseTasks from Ids
	T := func(tids ...string) Tasks {
		res := Tasks{}
		for _, tid := range tids {
			res = append(res, &testCaseTask{tid, 10, true})
		}
		return res
	}

	testCases := map[string]testCase{
		"test case 13*9 + 1*2": {
			input: WorkerTasks{
				"w1": T("t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9"),
				"w2": T("t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9"),
				"w3": T("t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9"),
				"w4": T("t10", "t11"),
			},
			want: WorkerTasks{
				"w1": T("t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9"),
				"w2": T("t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9", "t1"),
				"w3": T("t3", "t4", "t5", "t6", "t7", "t8", "t9", "t1", "t2"),
				"w4": T("t10", "t11"),
			},
		},
	}

	copts := cmp.Options{
		//cmpopts.IgnoreUnexported(testCaseTask{}),
	}

	for title, tc := range testCases {
		got := tc.input
		got.SortTasks()

		if diff := cmp.Diff(tc.want, got, copts); diff != "" {
			t.Errorf("%s: mismatch (-want +got):\n%s", title, diff)
		}
	}
}
