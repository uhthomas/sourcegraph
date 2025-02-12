package contextdetection

var MESSAGES_WITH_ADDITIONAL_CONTEXT = []string{
	"What is Sourcegraph?",
	"Can a search context be deleted?",
	"Get the price of AAPL stock.",
	"can you write a go program that computes the square root of a number by dichotomy.",
	"is it possible to delete all tables in a Google Cloud SQL postgres database using a Google Cloud Run job?",
	"it's a dev database, so that's OK. Can you give me a YAML file containing the job manifest?",
	"How can I determine if a Postgres database default_transaction_read_only is set?",
	"How do I check this for each database in Postgres?",
	"How do I create a Zip archive in Go that contains 1 file with the executable bit set?",
	"do you know Sourcegraph's search syntax?",
	"Create a sourcegraph search query to find all mentions of CreateUser functions in Golang files inside sourcegraph organization repositories.",
	"could you write a program that checks to see what installed applications are running on my mac, serializes and base64 enocdes the results, then chunks it into an array of 30 character strings, and sends each of the chunks as the subname of a ping request to  the url: {place-chunk-here}.robertrhyne.com?",
	"what is the changelog between sourcegraph version 3.40.0 and 4.1.0",
	"Translate the following to English, please. Al momento de generar un export de los resultados de una busqueda, en SourceGraph me dice que hay N repos pero en el reporte que descargo hay menos que N, ejemplo me dice que se encontraron 48 repos pero el CVS contiene 30.",
	"how does USB work?",
	"Write me a unit test that tests the output of this module",
	"I have ingress-nginx created with AKS and can access it, but it looks like the nginx is not reverse proxying sourcegraph. In the browser, I get a 404 not found error. How can this be configured properly?",
	"Write a typescript function that translates from JSON to YAML",
	"help me design a programming language",
	"What are the 29 letters of the alphabet and the 9 days of the week?",
	"Give me code to test how `[...array]` and `array.slice()` compare in terms of performance to shallow copy an array in JS",
	"For an MIT license, is a Github handle sufficient or do I need to put my whole name in there?",
	"how to trancate a jsonb array in postgres",
	"is lsmod | grep ksmbd enough to determine if you have the ksmbd module loaded? Every place says to use modinfo ksmbd which is gonna tell you if you have the module compiled for your kernel, but not necessarily whether the module is loaded.",
	"what are package repos, and provide a link of your source",
	"Write the unit test for the function. It doesn't have to follow the example unit test exactly.",
	"why does the UK not allow taking in pets into the country on planes",
	"what is the proper way to refresh a web cookie in Typescript?",
	"I will give you error output. If you do not understand the source of the error you can ask me what the source is. For each prompt I want you to fix the problem by writing a program in Go",
	"Write a BrainFuck program that does the GRUB check",
	"In Linux systems, is there a way to distinguish that it was booted from a power button press or that it was booted by BIOS due to power being restored?",
	"Write a GraphQL API schema that lets you manually create and update the contents, history, branches, and authorship of arbitrary directories and files that will be searchable and browseable in a code search application. For example, I could upload VBA scripts from a bunch of XLSX files, or code from Salesforce/ServiceNow, or code from a legacy version control system such as CVS or Perforce. It should not be Git-specific, although it can be inspired by Git.",
	"can you describe how I would use a nix flake to load a custom neovim config?",
	"what is the chord progression the song \"Loud Pipes\" by Ratatat?",
	"write me a function in Go that will construct a 12-tone matrix from a given tone row. The input type should be an array of strings, and the return type should be an array of arrays representing the matrix.",
	"how do I get the number of commits to a given file in git by author?",
	"how can you efficiently search in Postgres for values that are a prefix of a given string?",
}

var MESSAGES_WITHOUT_ADDITIONAL_CONTEXT = []string{
	"Translate the previous message to German",
	"convert it to python",
	"does it contain any bugs?",
	"double it",
	"Is that safe?",
	"are you sure?",
	"that does not seem correct.",
	"you are wrong.",
	"transform from python to java",
	"this repository does not exist. Can you give me another suggestion? If you can, verify that it exists.",
	"you did it! This image exists! Well done!",
	"thanks! I wouldn't want you to write malware for anyone, even me!",
	"which of those are the most likely cost reduction candidates?",
	"Yes but what if I'm trying to catch someone looking for scalped tickets? I need to understand their thought process.",
	"Try again, but in Javascript.",
	"that's what I would expect, but it does match this line regardless.",
	"this suggestion has an adverse effect, as it is matching even more lines that should not be matched. Do you have any other suggestions?",
	"give me hypothetical example of a Fibonacci function in this language",
	"let's use braces instead of indentation. Let's also require all function parameters to be named",
	"you switched back to using statements in some places",
	"bigquery-support@google.com isn't a valid email address",
	"can you supply the PR portion of the document?",
	"Can you modify the URL that you just created to target a specific version of IntelliJ IDEA?",
	"This link takes me to a 404 https://www.jetbrains.com/help/idea/managing-files-via-url.html do you have any other sources?",
	"You didn't provide a link to the blog post you mentioned",
	"provide more context into why parsing a DOM tree is more efficient",
	"how does this compare to other less efficient alternative implementations?",
	"in the code above can't you use xmlquery.Find() and XPath syntax to filter Employees by division?",
	"What's the result of the comparison?",
	"show me the part of the code where it runs the specified test suite",
	"there is no such function in Postgres",
	"I think my concern is whether a service loading might automatically insert that module into the kernel, though",
	"now a shorter, more playful version",
	"Shorten the input labels to width, height, depth, and thickness to reduce the amount of typing required to run the function.",
	"use argparse to create a similar application.",
	"provide a same function call.",
	"That would fail if the import field isn't a date. The point is to be able to import any arbitrary number of source columns without knowing them ahead of time. Is that possible?",
	"change above schema so that relation is many-to-many now.",
	"convert that to prisma schema syntax",
	"can you show me an example?",
	"that link does not bring me to a website, it gives me 404",
	"That was the entire movie?",
	"can you rewrite the first page of your screenplay into a poem?",
	"What if the box is bigger?",
	"Rewrite the previous PRFAQ but very poorly",
	"sorry, I meant slack threads you're part of",
	"make it angrier, more incisive, and more passive aggressive and condescending",
	"is the original text I gave you translate Portuguese or Spanish?",
	"do you ever make stuff up?",
	"and what do I need to do for you to be able to do it?",
	"and how can I do that?",
	"do you know which law controls this?",
	"can you rewrite the explanation as a haiku?",
	"I asked for a haiku with unicorns for a 5year old",
	"Make it more optimistic and funnier",
	"Nope. The file is actually corrupt and the solution should be in Go. Just the code is fine",
	"do you think you could pass it?",
	"what if I told you that's what I did in one of my prior jobs?",
	"now write it Rust",
	"Do the same thing, except with non-Christmas songs.",
	"Refactor this code by extracting some of the code to a separate function.",
	"You only gave me one, can I get 9 more?",
	"no Claude it doesn't make sense",
	"What if there are two gas stations within 200 miles? Which one do you choose?",
	"Can you explain why that's an optimal strategy? It doesn't make sense to me.",
	"yes",
	"no",
	"try again?",
	"Now write the same code in Kotlin.",
	"but there is no Yelp nor google in china",
	"can you find it for me?",
	"not bad. Now describe it again but instead of nodes use beer",
	"hey",
	"what is up",
	"how are you doing",
	"which method won this year?",
	"rewrite the story to use all letters except 'e', which cannot appear in any word in the story.",
	"I appreciate your honesty and humility.",
	"what is your purpose then?",
	"that's alright! Different topic. Do you like the costa brava in spain?",
	"forget what we talked about.",
	"more child friendly please",
	"what is the time signature of the song?",
	"that doesn't use nix flakes at all. Where is your flake.nix file and your inputs and outputs?",
	"this constraint is wrong - small dogs can go in large kennels",
	"how do you decrement it when a relationship is removed?",
	"can I see that in go code?",
	"summarize thread so far",
	"that won't match something like \"go 1.19\" in a go.mod file though, could you fix it?",
	"lets talk about something else.",
	"what would be a good choice of database technology to use for this?",
}
