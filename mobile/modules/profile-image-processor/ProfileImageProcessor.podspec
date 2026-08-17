Pod::Spec.new do |s|
  s.name           = 'ProfileImageProcessor'
  s.version        = '1.0.0'
  s.summary        = 'Bounded native profile image processing'
  s.description    = 'Metadata-first image validation and decode-time downsampling.'
  s.license        = { :type => 'MIT' }
  s.author         = 'Decorebator'
  s.homepage       = 'https://decorebator.com'
  s.platforms      = { :ios => '15.1' }
  s.swift_version  = '5.9'
  s.source         = { :git => 'https://github.com/decorebator/decorebator-v2.git' }
  s.static_framework = true
  s.dependency 'ExpoModulesCore'
  s.pod_target_xcconfig = { 'DEFINES_MODULE' => 'YES' }
  s.source_files = 'ios/**/*.{h,m,mm,swift,hpp,cpp}'
end
