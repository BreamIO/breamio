package eyetracker

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFilter(t *testing.T) {
	Convey("Filter should not return nil", t, func() {
		So(Filter(Point2D{}, Point2D{}), ShouldNotBeNil)
	})

	Convey("Filter should average the points given in both axis", t, func() {
		So(Filter(Point2D{-1, -1}, Point2D{1, 1}), ShouldResemble, Point2D{})
		So(Filter(Point2D{0.5, 0.5}, Point2D{0.5, 0.5}), ShouldResemble, Point2D{0.5, 0.5})
	})
}
