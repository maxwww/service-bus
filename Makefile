.PHONY: git

git:
	git tag $(tag) && git push origin && git push origin $(tag)

