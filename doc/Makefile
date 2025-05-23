STD_OPTS = \
	-V documentclass:article \
	-V header-includes:"\usepackage{fancyhdr} \pagestyle{fancy} \lhead{Handleiding Dock} \rhead{$(VERSION)}" \
	-V linestretch:1.25 \
	--number-sections \
	-f markdown \
	-t latex

%.pdf:	%.md *.png
	docker run --rm -v `pwd`:/source jagregory/pandoc $(STD_OPTS) \
		$< \
		-o $@
