package devicedeployment

import (
	"time"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

// IDMatches checks if the ID of the device dev matches the ID defined in the
// selector
func IDMatches(dev corev1beta1.Device, selector appsv1beta1.Selector) bool {
	did := selector.MatchByDeviceID
	if did == nil {
		return false
	}
	if dev.ObjectMeta.Name == *did {
		return true
	}
	return false
}

// LabelsMatch checks if the device dev matches the label selector defined in
// the selector
func LabelsMatch(dev corev1beta1.Device, selector appsv1beta1.Selector) bool {
	result := false
	for k, v := range selector.MatchByLabels {
		l, ok := dev.ObjectMeta.Labels[k]
		if !ok || l != v {
			return false
		}
		result = true
	}
	return result
}

// Matches checks if a selector matches the given device
func Matches(dev corev1beta1.Device, selector appsv1beta1.Selector) bool {
	return IDMatches(dev, selector) || LabelsMatch(dev, selector)
}

// precision returns the precision of a gridBoxDeployment which is currently
// defined as the number of conditions in the MatchByLabels selector
func precision(gbxd appsv1beta1.DeviceDeployment) int {
	if gbxd.Spec.Selector.MatchByLabels == nil {
		return 0
	}
	return len(gbxd.Spec.Selector.MatchByLabels)
}

// lastUpdatedAt returns the last updated time of the GridBoxDeployment
func lastUpdatedAt(gbxd appsv1beta1.DeviceDeployment) (time.Time, error) {
	return time.Parse(time.RFC3339, gbxd.Status.LastUpdatedAt)
}

// findMatchingDeployments returns all matching deployments of a the given list
// that match box gbx
func findMatchingDeployments(gbx corev1beta1.Device, gbxds []appsv1beta1.DeviceDeployment) map[string]appsv1beta1.DeviceDeployment {
	byApp := make(map[string]appsv1beta1.DeviceDeployment)
	for _, gbxd := range gbxds {
		if !Matches(gbx, gbxd.Spec.Selector) {
			continue
		}

		appName := gbxd.Spec.App

		curd, ok := byApp[appName]
		if !ok {
			byApp[appName] = gbxd
			continue
		}

		// NOTE: we currently use curd therefore the following will
		// only check of the cases where we need to change to gbxd
		if !IDMatches(gbx, curd.Spec.Selector) {
			if !IDMatches(gbx, gbxd.Spec.Selector) {
				// Both use labels for selection instead of IDs, so use the one that is
				// more precise
				gbxdPrec := precision(gbxd)
				curdPrec := precision(curd)

				switch {
				case gbxdPrec > curdPrec:
					// gbxd more precise => choose gbxd
					byApp[appName] = gbxd
				case gbxdPrec == curdPrec:
					// if both have equal precision use the
					// most recent, i.e. the one that was
					// most recently updated
					gbxdTs, err := lastUpdatedAt(gbxd)
					if err != nil {
						// Ignore error and leave curd
						// as deployment
					}
					curdTs, err := lastUpdatedAt(curd)
					if err != nil {
						// Cannot get ts of curd => use gbxd
						byApp[appName] = gbxd
						continue
					}
					// Can get ts of both => use most recent
					// one
					if gbxdTs.After(curdTs) {
						byApp[appName] = gbxd
					}
				}
			} else {
				// Use gbxd as it uses ID compared to curd
				// which uses labels
				byApp[appName] = gbxd
			}
		} else {
			if IDMatches(gbx, gbxd.Spec.Selector) {
				// Both use ID matching => use most recent one
				gbxdTs, err := lastUpdatedAt(gbxd)
				if err != nil {
					// Ignore error and leave curd as deployment
				}
				curdTs, err := lastUpdatedAt(curd)
				if err != nil {
					// Cannot get ts of curd => use gbxd
					byApp[appName] = gbxd
					continue
				}
				// Can get ts of both => use most recent
				// one
				if gbxdTs.After(curdTs) {
					byApp[appName] = gbxd
				}
			}
		}
	}

	return byApp
}

// findMatchingContainers returns all containers belonging to the given gridBox
// indexed by the app. While it is generally only allowed to have one container
// per app per box we return a list of containers for each app to allow to
// handle conflicts properly
func findMatchingContainers(box corev1beta1.Device, containers []corev1beta1.DevicePod) map[string][]corev1beta1.DevicePod {
	result := make(map[string][]corev1beta1.DevicePod)

	for _, c := range containers {
		if id := c.Spec.DeviceID; id != "" && id == box.Name {
			app := appsv1beta1.ExtractAppName(c)
			result[app] = append(result[app], c)
		}
	}
	return result
}
