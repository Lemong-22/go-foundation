package playground

type Command struct {
	ID          string  `json:"id"`
	Group       string  `json:"group"`
	Label       string  `json:"label"`
	Command     string  `json:"command"`
	Title       string  `json:"title"`
	About       string  `json:"about"`
	Description string  `json:"description"`
	Example     Example `json:"example"`
	Fields      []Field `json:"fields"`
}

type Example struct {
	Command string `json:"command"`
	Output  string `json:"output"`
}

type Field struct {
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	Kind         string   `json:"kind"`
	Required     bool     `json:"required"`
	RequiredWhen string   `json:"requiredWhen,omitempty"`
	VisibleWhen  string   `json:"visibleWhen,omitempty"`
	Default      string   `json:"default,omitempty"`
	Placeholder  string   `json:"placeholder,omitempty"`
	Options      []string `json:"options,omitempty"`
	Binding      Binding  `json:"binding"`
}

type Binding struct {
	Flag       string `json:"flag,omitempty"`
	Argument   bool   `json:"argument,omitempty"`
	BoolFlag   bool   `json:"boolFlag,omitempty"`
	AllowZero  bool   `json:"allowZero,omitempty"`
	AllowEmpty bool   `json:"allowEmpty,omitempty"`
	MultiValue bool   `json:"multiValue,omitempty"`
	Separator  string `json:"separator,omitempty"`
}

func Catalog() []Command {
	commands := []Command{
		{
			ID:          "course-create",
			Group:       "Course",
			Label:       "create",
			Command:     "course create",
			Title:       "Create Course",
			About:       "Create a draft course for an instructor.",
			Description: "Creates a course with a title, slug, optional description, and owning instructor id. The core service still validates slug and instructor id shape.",
			Example: Example{
				Command: `course-cli course create --title "Intro to Go" --slug intro-to-go --description "Learn Go" --instructor-id 550e8400-e29b-41d4-a716-446655440010`,
				Output:  "550e8400-e29b-41d4-a716-446655440000",
			},
			Fields: []Field{
				textFlag("title", "Title", "--title", true, "Intro to Go"),
				textFlag("slug", "Slug", "--slug", true, "intro-to-go"),
				textareaFlag("description", "Description", "--description", false, "Learn Go"),
				textFlag("instructor-id", "Instructor ID", "--instructor-id", false, "550e8400-e29b-41d4-a716-446655440010"),
			},
		},
		{
			ID:          "course-list",
			Group:       "Course",
			Label:       "list",
			Command:     "course list",
			Title:       "List Courses",
			About:       "List courses with optional lifecycle filtering.",
			Description: "Returns courses in table, JSON, or quiet id-only output. Status is optional and currently supports draft or published.",
			Example: Example{
				Command: `course-cli course list --status published --output json`,
				Output:  `[{"id":"550e8400-e29b-41d4-a716-446655440000","title":"Intro to Go","status":"published"}]`,
			},
			Fields: []Field{
				selectFlag("status", "Status", "--status", false, "", []string{"", "draft", "published"}),
				outputField(),
			},
		},
		{
			ID:          "course-get",
			Group:       "Course",
			Label:       "get",
			Command:     "course get",
			Title:       "Get Course",
			About:       "Fetch one course by id.",
			Description: "Loads course detail and renders it with the same table, JSON, or quiet output renderer used by the CLI.",
			Example: Example{
				Command: `course-cli course get 550e8400-e29b-41d4-a716-446655440000 --output table`,
				Output:  "FIELD  VALUE\nID     550e8400-e29b-41d4-a716-446655440000",
			},
			Fields: []Field{
				argField("course-id", "Course ID", true, "550e8400-e29b-41d4-a716-446655440000"),
				outputField(),
			},
		},
		{
			ID:          "course-update",
			Group:       "Course",
			Label:       "update",
			Command:     "course update",
			Title:       "Update Course",
			About:       "Change course metadata.",
			Description: "Updates only fields supplied in the form. Leave optional fields blank to keep the current value.",
			Example: Example{
				Command: `course-cli course update 550e8400-e29b-41d4-a716-446655440000 --title "Advanced Go"`,
				Output:  "550e8400-e29b-41d4-a716-446655440000",
			},
			Fields: []Field{
				argField("course-id", "Course ID", true, "550e8400-e29b-41d4-a716-446655440000"),
				textFlag("title", "Title", "--title", false, "Advanced Go"),
				textareaFlag("description", "Description", "--description", false, "Deeper course description"),
				textFlag("slug", "Slug", "--slug", false, "advanced-go"),
			},
		},
		{
			ID:          "course-delete",
			Group:       "Course",
			Label:       "delete",
			Command:     "course delete",
			Title:       "Delete Course",
			About:       "Delete a course and its lessons.",
			Description: "Runs the destructive course delete command. Use force to skip the CLI confirmation prompt in this browser flow.",
			Example: Example{
				Command: `course-cli course delete 550e8400-e29b-41d4-a716-446655440000 --force`,
				Output:  "course deleted",
			},
			Fields: []Field{
				argField("course-id", "Course ID", true, "550e8400-e29b-41d4-a716-446655440000"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "course-publish",
			Group:       "Course",
			Label:       "publish",
			Command:     "course publish",
			Title:       "Publish Course",
			About:       "Move a course to published status.",
			Description: "Publishes a draft course. The domain prevents publishing a course that is already published.",
			Example: Example{
				Command: `course-cli course publish 550e8400-e29b-41d4-a716-446655440000`,
				Output:  "course published",
			},
			Fields: []Field{argField("course-id", "Course ID", true, "550e8400-e29b-41d4-a716-446655440000")},
		},
		{
			ID:          "course-unpublish",
			Group:       "Course",
			Label:       "unpublish",
			Command:     "course unpublish",
			Title:       "Unpublish Course",
			About:       "Move a course back to draft status.",
			Description: "Unpublishes a published course. The domain rejects courses that are already draft.",
			Example: Example{
				Command: `course-cli course unpublish 550e8400-e29b-41d4-a716-446655440000`,
				Output:  "course unpublished",
			},
			Fields: []Field{argField("course-id", "Course ID", true, "550e8400-e29b-41d4-a716-446655440000")},
		},
		{
			ID:          "import-plan",
			Group:       "Import",
			Label:       "plan",
			Command:     "import plan",
			Title:       "Plan Import",
			About:       "Compute a deterministic import plan from a local zip path.",
			Description: "Runs the same import plan command used by automation. The browser import console supports direct zip upload; this command form is for server-local zip paths.",
			Example: Example{
				Command: `course-cli import plan ./course.zip --format json --instructor-id 550e8400-e29b-41d4-a716-446655440010`,
				Output:  `{"format_version":"1","operations":[],"conflicts":[]}`,
			},
			Fields: []Field{
				argField("zip-path", "Zip Path", true, "./course.zip"),
				selectFlag("format", "Format", "--format", false, "json", []string{"json", "table"}),
				textFlag("output", "Output File", "--output", false, "./plan.json"),
				textFlag("instructor-id", "Instructor ID", "--instructor-id", false, "550e8400-e29b-41d4-a716-446655440010"),
			},
		},
		{
			ID:          "import-apply",
			Group:       "Import",
			Label:       "apply",
			Command:     "import apply",
			Title:       "Apply Import",
			About:       "Apply a fresh or resolved import plan from a local zip path.",
			Description: "Runs the same import apply command used by automation. Use force to skip confirmation in the browser flow.",
			Example: Example{
				Command: `course-cli import apply ./course.zip --resolved-plan ./plan.json --conflict-strategy update --force --instructor-id 550e8400-e29b-41d4-a716-446655440010`,
				Output:  "FIELD  VALUE\nAPPLIED  4",
			},
			Fields: []Field{
				argField("zip-path", "Zip Path", true, "./course.zip"),
				selectFlag("format", "Format", "--format", false, "table", []string{"table", "json"}),
				textFlag("resolved-plan", "Resolved Plan", "--resolved-plan", false, "./plan.json"),
				selectFlag("conflict-strategy", "Conflict Strategy", "--conflict-strategy", false, "fail", []string{"fail", "skip", "update"}),
				boolFlag("force", "Force", "--force"),
				textFlag("instructor-id", "Instructor ID", "--instructor-id", false, "550e8400-e29b-41d4-a716-446655440010"),
			},
		},
		{
			ID:          "quiz-create",
			Group:       "Quiz Builder",
			Label:       "create",
			Command:     "quiz create",
			Title:       "Create Quiz",
			About:       "Create quiz metadata for a course.",
			Description: "Creates a quiz owned by one course. The quiz service validates the course id, title, and optional pass threshold.",
			Example: Example{
				Command: `course-cli quiz create --course-id 550e8400-e29b-41d4-a716-446655440000 --title "Go Basics Quiz" --pass-threshold 0.8`,
				Output:  "550e8400-e29b-41d4-a716-446655440040",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				textFlag("title", "Title", "--title", true, "Go Basics Quiz"),
				numberFlag("pass-threshold", "Pass Threshold", "--pass-threshold", false, "0.8"),
			},
		},
		{
			ID:          "quiz-list",
			Group:       "Quiz Builder",
			Label:       "list",
			Command:     "quiz list",
			Title:       "List Quizzes",
			About:       "List quizzes for one course.",
			Description: "Returns quiz metadata for the selected course using the same table, JSON, or quiet output renderer used by the CLI.",
			Example: Example{
				Command: `course-cli quiz list --course-id 550e8400-e29b-41d4-a716-446655440000 --output table`,
				Output:  "ID                                    COURSE_ID                             TITLE",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				outputField(),
			},
		},
		{
			ID:          "quiz-get",
			Group:       "Quiz Builder",
			Label:       "get",
			Command:     "quiz get",
			Title:       "Get Quiz",
			About:       "Fetch one quiz and its ordered questions.",
			Description: "Loads quiz detail through the quiz service port and renders the metadata plus current question sequence.",
			Example: Example{
				Command: `course-cli quiz get 550e8400-e29b-41d4-a716-446655440040 --output json`,
				Output:  `{"id":"550e8400-e29b-41d4-a716-446655440040","title":"Go Basics Quiz","questions":[]}`,
			},
			Fields: []Field{
				argField("quiz-id", "Quiz ID", true, "550e8400-e29b-41d4-a716-446655440040"),
				outputField(),
			},
		},
		{
			ID:          "quiz-update",
			Group:       "Quiz Builder",
			Label:       "update",
			Command:     "quiz update",
			Title:       "Update Quiz",
			About:       "Change quiz metadata.",
			Description: "Updates only the supplied quiz fields. Leave optional fields blank to keep the current value.",
			Example: Example{
				Command: `course-cli quiz update 550e8400-e29b-41d4-a716-446655440040 --title "Updated Quiz" --pass-threshold 0.75`,
				Output:  "550e8400-e29b-41d4-a716-446655440040",
			},
			Fields: []Field{
				argField("quiz-id", "Quiz ID", true, "550e8400-e29b-41d4-a716-446655440040"),
				textFlag("title", "Title", "--title", false, "Updated Quiz"),
				numberFlag("pass-threshold", "Pass Threshold", "--pass-threshold", false, "0.75"),
			},
		},
		{
			ID:          "quiz-delete",
			Group:       "Quiz Builder",
			Label:       "delete",
			Command:     "quiz delete",
			Title:       "Delete Quiz",
			About:       "Delete a quiz that is not embedded in lessons.",
			Description: "Runs the destructive quiz delete command. Use force to skip confirmation; the quiz service still rejects deletion when lessons reference the quiz.",
			Example: Example{
				Command: `course-cli quiz delete 550e8400-e29b-41d4-a716-446655440040 --force`,
				Output:  "quiz deleted",
			},
			Fields: []Field{
				argField("quiz-id", "Quiz ID", true, "550e8400-e29b-41d4-a716-446655440040"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "quiz-question-add",
			Group:       "Quiz Builder",
			Label:       "question add",
			Command:     "quiz question add",
			Title:       "Add Quiz Question",
			About:       "Append or insert a choice question.",
			Description: "Adds a single-choice or multiple-choice question through the quiz service. Enter answer options one per line and correct indices as comma-separated zero-based positions.",
			Example: Example{
				Command: `course-cli quiz question add --quiz-id 550e8400-e29b-41d4-a716-446655440040 --type single --prompt "Which keyword starts a goroutine?" --option go --option defer --correct 0 --explanation "go starts a new goroutine"`,
				Output:  "550e8400-e29b-41d4-a716-446655440050",
			},
			Fields: []Field{
				textFlag("quiz-id", "Quiz ID", "--quiz-id", true, "550e8400-e29b-41d4-a716-446655440040"),
				selectFlag("type", "Type", "--type", true, "single", []string{"single", "multiple"}),
				textareaFlag("prompt", "Prompt", "--prompt", true, "Which keyword starts a goroutine?"),
				multiValueTextareaFlag("option", "Options", "--option", true, "go\ndefer", "\n"),
				multiValueTextFlag("correct", "Correct Indices", "--correct", true, "0", ","),
				textareaFlag("explanation", "Explanation", "--explanation", false, "go starts a new goroutine"),
				numberFlag("position", "Position", "--position", false, "0"),
			},
		},
		{
			ID:          "quiz-question-list",
			Group:       "Quiz Builder",
			Label:       "question list",
			Command:     "quiz question list",
			Title:       "List Quiz Questions",
			About:       "List ordered questions for one quiz.",
			Description: "Returns the current question sequence for the selected quiz using the quiz service port.",
			Example: Example{
				Command: `course-cli quiz question list --quiz-id 550e8400-e29b-41d4-a716-446655440040 --output table`,
				Output:  "ID                                    QUIZ_ID                               POSITION  TYPE",
			},
			Fields: []Field{
				textFlag("quiz-id", "Quiz ID", "--quiz-id", true, "550e8400-e29b-41d4-a716-446655440040"),
				outputField(),
			},
		},
		{
			ID:          "quiz-question-get",
			Group:       "Quiz Builder",
			Label:       "question get",
			Command:     "quiz question get",
			Title:       "Get Quiz Question",
			About:       "Fetch one quiz question by id.",
			Description: "Loads a question and renders its prompt, ordered options, correct indices, and explanation.",
			Example: Example{
				Command: `course-cli quiz question get 550e8400-e29b-41d4-a716-446655440050 --output json`,
				Output:  `{"id":"550e8400-e29b-41d4-a716-446655440050","type":"single","correctIndices":[0]}`,
			},
			Fields: []Field{
				argField("question-id", "Question ID", true, "550e8400-e29b-41d4-a716-446655440050"),
				outputField(),
			},
		},
		{
			ID:          "quiz-question-update",
			Group:       "Quiz Builder",
			Label:       "question update",
			Command:     "quiz question update",
			Title:       "Update Quiz Question",
			About:       "Edit a question prompt, options, correct indices, or explanation.",
			Description: "Updates only fields supplied in the form. Options are submitted as repeated CLI flags after splitting one option per line.",
			Example: Example{
				Command: `course-cli quiz question update 550e8400-e29b-41d4-a716-446655440050 --prompt "Updated prompt" --option "Option A" --option "Option B" --correct 1 --explanation "Updated explanation"`,
				Output:  "550e8400-e29b-41d4-a716-446655440050",
			},
			Fields: []Field{
				argField("question-id", "Question ID", true, "550e8400-e29b-41d4-a716-446655440050"),
				textareaFlag("prompt", "Prompt", "--prompt", false, "Updated prompt"),
				multiValueTextareaFlag("option", "Options", "--option", false, "Option A\nOption B", "\n"),
				multiValueTextFlag("correct", "Correct Indices", "--correct", false, "1", ","),
				textareaFlag("explanation", "Explanation", "--explanation", false, "Updated explanation"),
			},
		},
		{
			ID:          "quiz-question-remove",
			Group:       "Quiz Builder",
			Label:       "question remove",
			Command:     "quiz question remove",
			Title:       "Remove Quiz Question",
			About:       "Remove one question from a quiz.",
			Description: "Runs the destructive question remove command. Use force to skip confirmation while the quiz service handles aggregate validation.",
			Example: Example{
				Command: `course-cli quiz question remove 550e8400-e29b-41d4-a716-446655440050 --force`,
				Output:  "quiz question removed",
			},
			Fields: []Field{
				argField("question-id", "Question ID", true, "550e8400-e29b-41d4-a716-446655440050"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "quiz-question-reorder",
			Group:       "Quiz Builder",
			Label:       "question reorder",
			Command:     "quiz question reorder",
			Title:       "Reorder Quiz Questions",
			About:       "Resequence questions inside one quiz.",
			Description: "Accepts comma-separated question-id:position pairs and applies the new order through the quiz aggregate.",
			Example: Example{
				Command: `course-cli quiz question reorder --quiz-id 550e8400-e29b-41d4-a716-446655440040 --order 550e8400-e29b-41d4-a716-446655440050:0,550e8400-e29b-41d4-a716-446655440051:1`,
				Output:  "quiz questions reordered",
			},
			Fields: []Field{
				textFlag("quiz-id", "Quiz ID", "--quiz-id", true, "550e8400-e29b-41d4-a716-446655440040"),
				textFlag("order", "Order", "--order", true, "550e8400-e29b-41d4-a716-446655440050:0,550e8400-e29b-41d4-a716-446655440051:1"),
			},
		},
		{
			ID:          "practice-create",
			Group:       "Practice Builder",
			Label:       "create",
			Command:     "practice create",
			Title:       "Create Practice",
			About:       "Create coding practice metadata and prompt content.",
			Description: "Creates a practice owned by one course. The practice service validates course existence, title, prompt, and language.",
			Example: Example{
				Command: `course-cli practice create --course-id 550e8400-e29b-41d4-a716-446655440000 --title "FizzBuzz" --language golang --prompt "Print fizz buzz" --starter-code "package main"`,
				Output:  "550e8400-e29b-41d4-a716-446655440060",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				textFlag("title", "Title", "--title", true, "FizzBuzz"),
				selectFlag("language", "Language", "--language", true, "golang", []string{"javascript", "typescript", "golang", "rust"}),
				textareaFlag("prompt", "Prompt", "--prompt", true, "Print fizz buzz"),
				allowEmpty(textareaFlag("starter-code", "Starter Code", "--starter-code", false, "package main")),
				allowEmpty(textareaFlag("solution", "Solution", "--solution", false, "fmt.Println()")),
			},
		},
		{
			ID:          "practice-list",
			Group:       "Practice Builder",
			Label:       "list",
			Command:     "practice list",
			Title:       "List Practices",
			About:       "List practices for one course.",
			Description: "Returns practice metadata for the selected course using the practice service port.",
			Example: Example{
				Command: `course-cli practice list --course-id 550e8400-e29b-41d4-a716-446655440000 --output table`,
				Output:  "ID                                    COURSE_ID                             TITLE",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				outputField(),
			},
		},
		{
			ID:          "practice-get",
			Group:       "Practice Builder",
			Label:       "get",
			Command:     "practice get",
			Title:       "Get Practice",
			About:       "Fetch one practice and its ordered test cases.",
			Description: "Loads a practice detail through the practice service port and renders prompt, source fields, and current test cases.",
			Example: Example{
				Command: `course-cli practice get 550e8400-e29b-41d4-a716-446655440060 --output json`,
				Output:  `{"id":"550e8400-e29b-41d4-a716-446655440060","title":"FizzBuzz","testCases":[]}`,
			},
			Fields: []Field{
				argField("practice-id", "Practice ID", true, "550e8400-e29b-41d4-a716-446655440060"),
				outputField(),
			},
		},
		{
			ID:          "practice-update",
			Group:       "Practice Builder",
			Label:       "update",
			Command:     "practice update",
			Title:       "Update Practice",
			About:       "Edit practice metadata, prompt, starter code, or solution.",
			Description: "Updates only fields supplied in the form. Starter code and solution can be deliberately cleared.",
			Example: Example{
				Command: `course-cli practice update 550e8400-e29b-41d4-a716-446655440060 --title "Updated FizzBuzz" --prompt "Updated prompt"`,
				Output:  "550e8400-e29b-41d4-a716-446655440060",
			},
			Fields: []Field{
				argField("practice-id", "Practice ID", true, "550e8400-e29b-41d4-a716-446655440060"),
				textFlag("title", "Title", "--title", false, "Updated FizzBuzz"),
				textareaFlag("prompt", "Prompt", "--prompt", false, "Updated prompt"),
				allowEmpty(textareaFlag("starter-code", "Starter Code", "--starter-code", false, "package main")),
				allowEmpty(textareaFlag("solution", "Solution", "--solution", false, "fmt.Println()")),
			},
		},
		{
			ID:          "practice-delete",
			Group:       "Practice Builder",
			Label:       "delete",
			Command:     "practice delete",
			Title:       "Delete Practice",
			About:       "Delete a practice that is not embedded in lessons.",
			Description: "Runs the destructive practice delete command. Use force to skip confirmation; the practice service still rejects deletion when lessons reference the practice.",
			Example: Example{
				Command: `course-cli practice delete 550e8400-e29b-41d4-a716-446655440060 --force`,
				Output:  "practice deleted",
			},
			Fields: []Field{
				argField("practice-id", "Practice ID", true, "550e8400-e29b-41d4-a716-446655440060"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "practice-testcase-add",
			Group:       "Practice Builder",
			Label:       "testcase add",
			Command:     "practice testcase add",
			Title:       "Add Practice Test Case",
			About:       "Append or insert one stdin/stdout test case.",
			Description: "Adds a test case through the practice service. Stdin, expected stdout, and name may be empty; position is optional.",
			Example: Example{
				Command: `course-cli practice testcase add --practice-id 550e8400-e29b-41d4-a716-446655440060 --stdin "3" --expected-stdout "Fizz" --name "multiple of three" --position 0`,
				Output:  "550e8400-e29b-41d4-a716-446655440070",
			},
			Fields: []Field{
				textFlag("practice-id", "Practice ID", "--practice-id", true, "550e8400-e29b-41d4-a716-446655440060"),
				allowEmpty(textareaFlag("stdin", "Stdin", "--stdin", false, "3")),
				allowEmpty(textareaFlag("expected-stdout", "Expected Stdout", "--expected-stdout", false, "Fizz")),
				allowEmpty(textFlag("name", "Name", "--name", false, "multiple of three")),
				numberFlag("position", "Position", "--position", false, "0"),
			},
		},
		{
			ID:          "practice-testcase-list",
			Group:       "Practice Builder",
			Label:       "testcase list",
			Command:     "practice testcase list",
			Title:       "List Practice Test Cases",
			About:       "List ordered test cases for one practice.",
			Description: "Returns the current test case sequence for the selected practice.",
			Example: Example{
				Command: `course-cli practice testcase list --practice-id 550e8400-e29b-41d4-a716-446655440060 --output table`,
				Output:  "ID                                    PRACTICE_ID                           POSITION",
			},
			Fields: []Field{
				textFlag("practice-id", "Practice ID", "--practice-id", true, "550e8400-e29b-41d4-a716-446655440060"),
				outputField(),
			},
		},
		{
			ID:          "practice-testcase-get",
			Group:       "Practice Builder",
			Label:       "testcase get",
			Command:     "practice testcase get",
			Title:       "Get Practice Test Case",
			About:       "Fetch one practice test case by id.",
			Description: "Loads a test case through its owning practice and renders stdin, expected stdout, name, and position.",
			Example: Example{
				Command: `course-cli practice testcase get 550e8400-e29b-41d4-a716-446655440070 --output json`,
				Output:  `{"id":"550e8400-e29b-41d4-a716-446655440070","expectedStdout":"Fizz"}`,
			},
			Fields: []Field{
				argField("testcase-id", "Test Case ID", true, "550e8400-e29b-41d4-a716-446655440070"),
				outputField(),
			},
		},
		{
			ID:          "practice-testcase-update",
			Group:       "Practice Builder",
			Label:       "testcase update",
			Command:     "practice testcase update",
			Title:       "Update Practice Test Case",
			About:       "Edit stdin, expected stdout, or label for one test case.",
			Description: "Updates only fields supplied in the form. Empty stdin, expected stdout, and name values are valid when deliberately submitted.",
			Example: Example{
				Command: `course-cli practice testcase update 550e8400-e29b-41d4-a716-446655440070 --stdin "5" --expected-stdout "Buzz" --name "multiple of five"`,
				Output:  "550e8400-e29b-41d4-a716-446655440070",
			},
			Fields: []Field{
				argField("testcase-id", "Test Case ID", true, "550e8400-e29b-41d4-a716-446655440070"),
				allowEmpty(textareaFlag("stdin", "Stdin", "--stdin", false, "5")),
				allowEmpty(textareaFlag("expected-stdout", "Expected Stdout", "--expected-stdout", false, "Buzz")),
				allowEmpty(textFlag("name", "Name", "--name", false, "multiple of five")),
			},
		},
		{
			ID:          "practice-testcase-remove",
			Group:       "Practice Builder",
			Label:       "testcase remove",
			Command:     "practice testcase remove",
			Title:       "Remove Practice Test Case",
			About:       "Remove one test case from a practice.",
			Description: "Runs the destructive test case remove command. Use force to skip confirmation while the practice service handles aggregate validation.",
			Example: Example{
				Command: `course-cli practice testcase remove 550e8400-e29b-41d4-a716-446655440070 --force`,
				Output:  "practice test case removed",
			},
			Fields: []Field{
				argField("testcase-id", "Test Case ID", true, "550e8400-e29b-41d4-a716-446655440070"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "practice-testcase-reorder",
			Group:       "Practice Builder",
			Label:       "testcase reorder",
			Command:     "practice testcase reorder",
			Title:       "Reorder Practice Test Cases",
			About:       "Resequence test cases inside one practice.",
			Description: "Accepts comma-separated testcase-id:position pairs and applies the new order through the practice aggregate.",
			Example: Example{
				Command: `course-cli practice testcase reorder --practice-id 550e8400-e29b-41d4-a716-446655440060 --order 550e8400-e29b-41d4-a716-446655440070:0,550e8400-e29b-41d4-a716-446655440071:1`,
				Output:  "practice test cases reordered",
			},
			Fields: []Field{
				textFlag("practice-id", "Practice ID", "--practice-id", true, "550e8400-e29b-41d4-a716-446655440060"),
				textFlag("order", "Order", "--order", true, "550e8400-e29b-41d4-a716-446655440070:0,550e8400-e29b-41d4-a716-446655440071:1"),
			},
		},
		{
			ID:          "test-create",
			Group:       "Test Builder",
			Label:       "create",
			Command:     "test create",
			Title:       "Create Test",
			About:       "Create test metadata for a course.",
			Description: "Creates a test owned by one course. The test service validates course existence, title, optional time limit, and pass threshold.",
			Example: Example{
				Command: `course-cli test create --course-id 550e8400-e29b-41d4-a716-446655440000 --title "Final Test" --time-limit-minutes 45 --pass-threshold 0.8`,
				Output:  "550e8400-e29b-41d4-a716-446655440080",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				textFlag("title", "Title", "--title", true, "Final Test"),
				numberFlag("time-limit-minutes", "Time Limit Minutes", "--time-limit-minutes", false, "45"),
				numberFlag("pass-threshold", "Pass Threshold", "--pass-threshold", false, "0.8"),
			},
		},
		{
			ID:          "test-list",
			Group:       "Test Builder",
			Label:       "list",
			Command:     "test list",
			Title:       "List Tests",
			About:       "List tests for one course.",
			Description: "Returns test metadata for the selected course using the test service port.",
			Example: Example{
				Command: `course-cli test list --course-id 550e8400-e29b-41d4-a716-446655440000 --output table`,
				Output:  "ID                                    COURSE_ID                             TITLE",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				outputField(),
			},
		},
		{
			ID:          "test-get",
			Group:       "Test Builder",
			Label:       "get",
			Command:     "test get",
			Title:       "Get Test",
			About:       "Fetch one test and its ordered items.",
			Description: "Loads test detail through the test service port and renders metadata, solution package fields, and current item sequence.",
			Example: Example{
				Command: `course-cli test get 550e8400-e29b-41d4-a716-446655440080 --output json`,
				Output:  `{"id":"550e8400-e29b-41d4-a716-446655440080","title":"Final Test","items":[]}`,
			},
			Fields: []Field{
				argField("test-id", "Test ID", true, "550e8400-e29b-41d4-a716-446655440080"),
				outputField(),
			},
		},
		{
			ID:          "test-update",
			Group:       "Test Builder",
			Label:       "update",
			Command:     "test update",
			Title:       "Update Test",
			About:       "Edit test metadata or solution package references.",
			Description: "Updates only supplied fields. The solution package fields are submitted together as one group for zip and explanation video references.",
			Example: Example{
				Command: `course-cli test update 550e8400-e29b-41d4-a716-446655440080 --title "Updated Test" --solution-zip-provider url --solution-zip-locator https://example.com/solution.zip --solution-video-provider url --solution-video-locator https://example.com/video.mp4 --solution-video-caption "Walkthrough"`,
				Output:  "550e8400-e29b-41d4-a716-446655440080",
			},
			Fields: []Field{
				argField("test-id", "Test ID", true, "550e8400-e29b-41d4-a716-446655440080"),
				textFlag("title", "Title", "--title", false, "Updated Test"),
				numberFlag("time-limit-minutes", "Time Limit Minutes", "--time-limit-minutes", false, "0"),
				numberFlag("pass-threshold", "Pass Threshold", "--pass-threshold", false, "0.9"),
				textFlag("solution-zip-provider", "Solution Zip Provider", "--solution-zip-provider", false, "url"),
				textFlag("solution-zip-locator", "Solution Zip Locator", "--solution-zip-locator", false, "https://example.com/solution.zip"),
				textFlag("solution-video-provider", "Solution Video Provider", "--solution-video-provider", false, "url"),
				textFlag("solution-video-locator", "Solution Video Locator", "--solution-video-locator", false, "https://example.com/video.mp4"),
				allowEmpty(textFlag("solution-video-caption", "Solution Video Caption", "--solution-video-caption", false, "Walkthrough")),
			},
		},
		{
			ID:          "test-delete",
			Group:       "Test Builder",
			Label:       "delete",
			Command:     "test delete",
			Title:       "Delete Test",
			About:       "Delete a test.",
			Description: "Runs the destructive test delete command. Use force to skip the CLI confirmation prompt in this browser flow.",
			Example: Example{
				Command: `course-cli test delete 550e8400-e29b-41d4-a716-446655440080 --force`,
				Output:  "test deleted",
			},
			Fields: []Field{
				argField("test-id", "Test ID", true, "550e8400-e29b-41d4-a716-446655440080"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "test-item-add",
			Group:       "Test Builder",
			Label:       "item add",
			Command:     "test item add",
			Title:       "Add Test Item",
			About:       "Append or insert a choice or coding item.",
			Description: "Adds a choice or coding item through the test service. Choice options are entered one per line; coding test cases use stdin::expected[::name] lines.",
			Example: Example{
				Command: `course-cli test item add --test-id 550e8400-e29b-41d4-a716-446655440080 --kind choice --prompt "Pick two" --type multiple --option A --option B --correct 0 --correct 1 --explanation "A and B"`,
				Output:  "550e8400-e29b-41d4-a716-446655440090",
			},
			Fields: []Field{
				textFlag("test-id", "Test ID", "--test-id", true, "550e8400-e29b-41d4-a716-446655440080"),
				selectFlag("kind", "Kind", "--kind", true, "choice", []string{"choice", "coding"}),
				textareaFlag("prompt", "Prompt", "--prompt", true, "Pick two"),
				requiredWhen(selectFlag("type", "Choice Type", "--type", false, "single", []string{"single", "multiple"}), "kind=choice"),
				requiredWhen(multiValueTextareaFlag("option", "Options", "--option", false, "A\nB", "\n"), "kind=choice"),
				requiredWhen(multiValueTextFlag("correct", "Correct Indices", "--correct", false, "0,1", ","), "kind=choice"),
				visibleWhen(allowEmpty(textareaFlag("explanation", "Explanation", "--explanation", false, "A and B")), "kind=choice"),
				requiredWhen(selectFlag("language", "Language", "--language", false, "golang", []string{"javascript", "typescript", "golang", "rust"}), "kind=coding"),
				visibleWhen(allowEmpty(textareaFlag("starter-code", "Starter Code", "--starter-code", false, "package main")), "kind=coding"),
				visibleWhen(allowEmpty(textareaFlag("solution", "Solution", "--solution", false, "func main() {}")), "kind=coding"),
				requiredWhen(multiValueTextareaFlag("testcase", "Test Cases", "--testcase", false, "1::1::sample\n::ok", "\n"), "kind=coding"),
				numberFlag("position", "Position", "--position", false, "0"),
			},
		},
		{
			ID:          "test-item-list",
			Group:       "Test Builder",
			Label:       "item list",
			Command:     "test item list",
			Title:       "List Test Items",
			About:       "List ordered items for one test.",
			Description: "Returns the current test item sequence with choice and coding summaries for the selected test.",
			Example: Example{
				Command: `course-cli test item list --test-id 550e8400-e29b-41d4-a716-446655440080 --output table`,
				Output:  "ID                                    TEST_ID                               POSITION  KIND",
			},
			Fields: []Field{
				textFlag("test-id", "Test ID", "--test-id", true, "550e8400-e29b-41d4-a716-446655440080"),
				outputField(),
			},
		},
		{
			ID:          "test-item-get",
			Group:       "Test Builder",
			Label:       "item get",
			Command:     "test item get",
			Title:       "Get Test Item",
			About:       "Fetch full detail for one test item.",
			Description: "Loads a selected test item and renders its full choice or coding payload.",
			Example: Example{
				Command: `course-cli test item get 550e8400-e29b-41d4-a716-446655440090 --output json`,
				Output:  `{"id":"550e8400-e29b-41d4-a716-446655440090","kind":"choice","choiceOptions":["A","B"]}`,
			},
			Fields: []Field{
				argField("item-id", "Item ID", true, "550e8400-e29b-41d4-a716-446655440090"),
				outputField(),
			},
		},
		{
			ID:          "test-item-update",
			Group:       "Test Builder",
			Label:       "item update",
			Command:     "test item update",
			Title:       "Update Test Item",
			About:       "Edit choice or coding item payload fields.",
			Description: "Updates only supplied item fields. Options and coding test cases are submitted as repeated flags after splitting one entry per line.",
			Example: Example{
				Command: `course-cli test item update 550e8400-e29b-41d4-a716-446655440090 --prompt "Updated prompt" --option A --option B --correct 1 --explanation ""`,
				Output:  "550e8400-e29b-41d4-a716-446655440090",
			},
			Fields: []Field{
				argField("item-id", "Item ID", true, "550e8400-e29b-41d4-a716-446655440090"),
				controlField("payload-kind", "Payload Kind", "choice", []string{"choice", "coding"}),
				textareaFlag("prompt", "Prompt", "--prompt", false, "Updated prompt"),
				visibleWhen(selectFlag("type", "Choice Type", "--type", false, "", []string{"", "single", "multiple"}), "payload-kind=choice"),
				visibleWhen(multiValueTextareaFlag("option", "Options", "--option", false, "A\nB", "\n"), "payload-kind=choice"),
				visibleWhen(multiValueTextFlag("correct", "Correct Indices", "--correct", false, "1", ","), "payload-kind=choice"),
				visibleWhen(allowEmpty(textareaFlag("explanation", "Explanation", "--explanation", false, "")), "payload-kind=choice"),
				visibleWhen(selectFlag("language", "Language", "--language", false, "", []string{"", "javascript", "typescript", "golang", "rust"}), "payload-kind=coding"),
				visibleWhen(allowEmpty(textareaFlag("starter-code", "Starter Code", "--starter-code", false, "")), "payload-kind=coding"),
				visibleWhen(allowEmpty(textareaFlag("solution", "Solution", "--solution", false, "updated solution")), "payload-kind=coding"),
				visibleWhen(multiValueTextareaFlag("testcase", "Test Cases", "--testcase", false, "stdin::stdout::case", "\n"), "payload-kind=coding"),
			},
		},
		{
			ID:          "test-item-remove",
			Group:       "Test Builder",
			Label:       "item remove",
			Command:     "test item remove",
			Title:       "Remove Test Item",
			About:       "Remove one item from a test.",
			Description: "Runs the destructive item remove command. Use force to skip confirmation while the test service handles aggregate validation.",
			Example: Example{
				Command: `course-cli test item remove 550e8400-e29b-41d4-a716-446655440090 --force`,
				Output:  "test item removed",
			},
			Fields: []Field{
				argField("item-id", "Item ID", true, "550e8400-e29b-41d4-a716-446655440090"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "test-item-reorder",
			Group:       "Test Builder",
			Label:       "item reorder",
			Command:     "test item reorder",
			Title:       "Reorder Test Items",
			About:       "Resequence items inside one test.",
			Description: "Accepts comma-separated item-id:position pairs and applies the new order through the test aggregate.",
			Example: Example{
				Command: `course-cli test item reorder --test-id 550e8400-e29b-41d4-a716-446655440080 --order 550e8400-e29b-41d4-a716-446655440090:1,550e8400-e29b-41d4-a716-446655440091:0`,
				Output:  "test items reordered",
			},
			Fields: []Field{
				textFlag("test-id", "Test ID", "--test-id", true, "550e8400-e29b-41d4-a716-446655440080"),
				textFlag("order", "Order", "--order", true, "550e8400-e29b-41d4-a716-446655440090:1,550e8400-e29b-41d4-a716-446655440091:0"),
			},
		},
		{
			ID:          "lesson-create",
			Group:       "Lesson",
			Label:       "create",
			Command:     "lesson create",
			Title:       "Create Lesson",
			About:       "Add a lesson to a course.",
			Description: "Creates a lesson in the selected course. Order is optional; when omitted, the usecase appends the lesson by default.",
			Example: Example{
				Command: `course-cli lesson create --course-id 550e8400-e29b-41d4-a716-446655440000 --title "First Lesson" --order 0`,
				Output:  "550e8400-e29b-41d4-a716-446655440020",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				textFlag("title", "Title", "--title", true, "First Lesson"),
				numberFlag("order", "Order", "--order", false, "0"),
			},
		},
		{
			ID:          "lesson-list",
			Group:       "Lesson",
			Label:       "list",
			Command:     "lesson list",
			Title:       "List Lessons",
			About:       "List lessons for one course.",
			Description: "Returns all lessons within a course ordered by lesson position.",
			Example: Example{
				Command: `course-cli lesson list --course-id 550e8400-e29b-41d4-a716-446655440000 --output table`,
				Output:  "ID                                    COURSE_ID                             ORDER  TITLE",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				outputField(),
			},
		},
		{
			ID:          "lesson-get",
			Group:       "Lesson",
			Label:       "get",
			Command:     "lesson get",
			Title:       "Get Lesson",
			About:       "Fetch one lesson by id.",
			Description: "Loads lesson detail and renders it with the same table, JSON, or quiet output renderer used by the CLI.",
			Example: Example{
				Command: `course-cli lesson get 550e8400-e29b-41d4-a716-446655440020 --output json`,
				Output:  `{"id":"550e8400-e29b-41d4-a716-446655440020","title":"First Lesson"}`,
			},
			Fields: []Field{
				argField("lesson-id", "Lesson ID", true, "550e8400-e29b-41d4-a716-446655440020"),
				outputField(),
			},
		},
		{
			ID:          "lesson-update",
			Group:       "Lesson",
			Label:       "update",
			Command:     "lesson update",
			Title:       "Update Lesson",
			About:       "Change lesson title.",
			Description: "Updates the lesson title when supplied in the form.",
			Example: Example{
				Command: `course-cli lesson update 550e8400-e29b-41d4-a716-446655440020 --title "Updated Lesson"`,
				Output:  "550e8400-e29b-41d4-a716-446655440020",
			},
			Fields: []Field{
				argField("lesson-id", "Lesson ID", true, "550e8400-e29b-41d4-a716-446655440020"),
				textFlag("title", "Title", "--title", false, "Updated Lesson"),
			},
		},
		{
			ID:          "lesson-delete",
			Group:       "Lesson",
			Label:       "delete",
			Command:     "lesson delete",
			Title:       "Delete Lesson",
			About:       "Remove one lesson.",
			Description: "Runs the destructive lesson delete command. Use force to skip the CLI confirmation prompt in this browser flow.",
			Example: Example{
				Command: `course-cli lesson delete 550e8400-e29b-41d4-a716-446655440020 --force`,
				Output:  "lesson deleted",
			},
			Fields: []Field{
				argField("lesson-id", "Lesson ID", true, "550e8400-e29b-41d4-a716-446655440020"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "lesson-reorder",
			Group:       "Lesson",
			Label:       "reorder",
			Command:     "lesson reorder",
			Title:       "Reorder Lessons",
			About:       "Resequence lessons in a course.",
			Description: "Accepts comma-separated lesson-id:position pairs and hands them to the lesson reorder usecase through the CLI parser.",
			Example: Example{
				Command: `course-cli lesson reorder --course-id 550e8400-e29b-41d4-a716-446655440000 --order 550e8400-e29b-41d4-a716-446655440020:0`,
				Output:  "lessons reordered",
			},
			Fields: []Field{
				textFlag("course-id", "Course ID", "--course-id", true, "550e8400-e29b-41d4-a716-446655440000"),
				textFlag("order", "Order", "--order", true, "550e8400-e29b-41d4-a716-446655440020:0,550e8400-e29b-41d4-a716-446655440021:1"),
			},
		},
		{
			ID:          "lesson-block-add",
			Group:       "Lesson Block",
			Label:       "add",
			Command:     "lesson block add",
			Title:       "Add Lesson Block",
			About:       "Append or insert text, video, quiz, and practice content blocks.",
			Description: "Creates an ordered block inside an existing lesson. Text blocks require markdown text; video blocks require provider and locator values; quiz and practice blocks require ids validated by the lesson usecase.",
			Example: Example{
				Command: `course-cli lesson block add --lesson-id 550e8400-e29b-41d4-a716-446655440020 --kind text --text "## Intro" --position 0`,
				Output:  "550e8400-e29b-41d4-a716-446655440030",
			},
			Fields: []Field{
				textFlag("lesson-id", "Lesson ID", "--lesson-id", true, "550e8400-e29b-41d4-a716-446655440020"),
				selectFlag("kind", "Kind", "--kind", true, "text", []string{"text", "video", "quiz", "practice"}),
				requiredWhen(textareaFlag("text", "Text Markdown", "--text", false, "## Intro"), "kind=text"),
				requiredWhen(textFlag("video-provider", "Video Provider", "--video-provider", false, "youtube"), "kind=video"),
				requiredWhen(textFlag("video-locator", "Video Locator", "--video-locator", false, "dQw4w9WgXcQ"), "kind=video"),
				visibleWhen(textFlag("video-caption", "Video Caption", "--video-caption", false, "Intro video"), "kind=video"),
				requiredWhen(textFlag("quiz-id", "Quiz ID", "--quiz-id", false, "550e8400-e29b-41d4-a716-446655440040"), "kind=quiz"),
				requiredWhen(textFlag("practice-id", "Practice ID", "--practice-id", false, "550e8400-e29b-41d4-a716-446655440060"), "kind=practice"),
				numberFlag("position", "Position", "--position", false, "0"),
			},
		},
		{
			ID:          "lesson-block-list",
			Group:       "Lesson Block",
			Label:       "list",
			Command:     "lesson block list",
			Title:       "List Lesson Blocks",
			About:       "List ordered blocks for a lesson.",
			Description: "Returns the text, video, quiz, and practice block sequence for one lesson using the same table, JSON, or quiet renderers as the CLI.",
			Example: Example{
				Command: `course-cli lesson block list --lesson-id 550e8400-e29b-41d4-a716-446655440020 --output table`,
				Output:  "ID                                    LESSON_ID                             POSITION  KIND",
			},
			Fields: []Field{
				textFlag("lesson-id", "Lesson ID", "--lesson-id", true, "550e8400-e29b-41d4-a716-446655440020"),
				outputField(),
			},
		},
		{
			ID:          "lesson-block-get",
			Group:       "Lesson Block",
			Label:       "get",
			Command:     "lesson block get",
			Title:       "Get Lesson Block",
			About:       "Fetch one block by id.",
			Description: "Loads a single content block and renders its typed payload fields. The usecase resolves the owning lesson aggregate from the block id.",
			Example: Example{
				Command: `course-cli lesson block get 550e8400-e29b-41d4-a716-446655440030 --output json`,
				Output:  `{"id":"550e8400-e29b-41d4-a716-446655440030","kind":"text","position":0}`,
			},
			Fields: []Field{
				argField("block-id", "Block ID", true, "550e8400-e29b-41d4-a716-446655440030"),
				outputField(),
			},
		},
		{
			ID:          "lesson-block-update",
			Group:       "Lesson Block",
			Label:       "update",
			Command:     "lesson block update",
			Title:       "Update Lesson Block",
			About:       "Edit a block payload without changing its kind.",
			Description: "Updates only supplied payload flags. Text blocks accept markdown text; video blocks accept provider, locator, and caption changes. Quiz and practice refs are changed by removing and adding a block.",
			Example: Example{
				Command: `course-cli lesson block update 550e8400-e29b-41d4-a716-446655440030 --text "Updated markdown"`,
				Output:  "550e8400-e29b-41d4-a716-446655440030",
			},
			Fields: []Field{
				argField("block-id", "Block ID", true, "550e8400-e29b-41d4-a716-446655440030"),
				textareaFlag("text", "Text Markdown", "--text", false, "Updated markdown"),
				textFlag("video-provider", "Video Provider", "--video-provider", false, "youtube"),
				textFlag("video-locator", "Video Locator", "--video-locator", false, "dQw4w9WgXcQ"),
				textFlag("video-caption", "Video Caption", "--video-caption", false, "Intro video"),
			},
		},
		{
			ID:          "lesson-block-remove",
			Group:       "Lesson Block",
			Label:       "remove",
			Command:     "lesson block remove",
			Title:       "Remove Lesson Block",
			About:       "Remove one block from a lesson.",
			Description: "Runs the destructive block remove command. Use force to skip the CLI confirmation prompt in this browser flow.",
			Example: Example{
				Command: `course-cli lesson block remove 550e8400-e29b-41d4-a716-446655440030 --force`,
				Output:  "lesson block removed",
			},
			Fields: []Field{
				argField("block-id", "Block ID", true, "550e8400-e29b-41d4-a716-446655440030"),
				boolFlag("force", "Force", "--force"),
			},
		},
		{
			ID:          "lesson-block-reorder",
			Group:       "Lesson Block",
			Label:       "reorder",
			Command:     "lesson block reorder",
			Title:       "Reorder Lesson Blocks",
			About:       "Resequence blocks inside one lesson.",
			Description: "Accepts comma-separated block-id:position pairs and applies the new order through the lesson aggregate.",
			Example: Example{
				Command: `course-cli lesson block reorder --lesson-id 550e8400-e29b-41d4-a716-446655440020 --order 550e8400-e29b-41d4-a716-446655440030:0,550e8400-e29b-41d4-a716-446655440031:1`,
				Output:  "lesson blocks reordered",
			},
			Fields: []Field{
				textFlag("lesson-id", "Lesson ID", "--lesson-id", true, "550e8400-e29b-41d4-a716-446655440020"),
				textFlag("order", "Order", "--order", true, "550e8400-e29b-41d4-a716-446655440030:0,550e8400-e29b-41d4-a716-446655440031:1"),
			},
		},
	}

	return commands
}

func textFlag(key string, label string, flag string, required bool, placeholder string) Field {
	return Field{
		Key:         key,
		Label:       label,
		Kind:        "text",
		Required:    required,
		Placeholder: placeholder,
		Binding:     Binding{Flag: flag},
	}
}

func textareaFlag(key string, label string, flag string, required bool, placeholder string) Field {
	field := textFlag(key, label, flag, required, placeholder)
	field.Kind = "textarea"
	return field
}

func multiValueTextFlag(key string, label string, flag string, required bool, placeholder string, separator string) Field {
	field := textFlag(key, label, flag, required, placeholder)
	field.Binding.MultiValue = true
	field.Binding.Separator = separator
	return field
}

func multiValueTextareaFlag(key string, label string, flag string, required bool, placeholder string, separator string) Field {
	field := textareaFlag(key, label, flag, required, placeholder)
	field.Binding.MultiValue = true
	field.Binding.Separator = separator
	return field
}

func numberFlag(key string, label string, flag string, required bool, placeholder string) Field {
	field := textFlag(key, label, flag, required, placeholder)
	field.Kind = "number"
	field.Binding.AllowZero = true
	return field
}

func selectFlag(key string, label string, flag string, required bool, defaultValue string, options []string) Field {
	return Field{
		Key:      key,
		Label:    label,
		Kind:     "select",
		Required: required,
		Default:  defaultValue,
		Options:  options,
		Binding:  Binding{Flag: flag},
	}
}

func controlField(key string, label string, defaultValue string, options []string) Field {
	return Field{
		Key:     key,
		Label:   label,
		Kind:    "select",
		Default: defaultValue,
		Options: options,
	}
}

func boolFlag(key string, label string, flag string) Field {
	return Field{
		Key:     key,
		Label:   label,
		Kind:    "checkbox",
		Binding: Binding{Flag: flag, BoolFlag: true},
	}
}

func requiredWhen(field Field, condition string) Field {
	field.RequiredWhen = condition
	field.VisibleWhen = condition
	return field
}

func visibleWhen(field Field, condition string) Field {
	field.VisibleWhen = condition
	return field
}

func allowEmpty(field Field) Field {
	field.Binding.AllowEmpty = true
	return field
}

func argField(key string, label string, required bool, placeholder string) Field {
	return Field{
		Key:         key,
		Label:       label,
		Kind:        "text",
		Required:    required,
		Placeholder: placeholder,
		Binding:     Binding{Argument: true},
	}
}

func outputField() Field {
	return selectFlag("output", "Output", "--output", false, "table", []string{"table", "json", "quiet"})
}
